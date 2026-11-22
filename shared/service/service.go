// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/observability"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func NewLogger(serviceName string) *slog.LevelVar {
	logger, level := newLogger(serviceName, os.Stdout)
	slog.SetDefault(logger)
	return level
}

func newLogger(serviceName string, output io.Writer) (*slog.Logger, *slog.LevelVar) {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	logger := slog.New(
		traceHandler{Handler: slog.NewJSONHandler(
			output,
			&slog.HandlerOptions{Level: level},
		)},
	).With("service", serviceName)

	for _, field := range []struct {
		env string
		key string
	}{
		{env: "SERVICE_VERSION", key: "service_version"},
		{env: "DEPLOYMENT_ENVIRONMENT", key: "deployment_environment"},
		{env: "POD_UID", key: "service_instance_id"},
		{env: "POD_NAME", key: "k8s_pod_name"},
		{env: "POD_NAMESPACE", key: "k8s_namespace_name"},
		{env: "NODE_NAME", key: "k8s_node_name"},
	} {
		if value := os.Getenv(field.env); value != "" {
			logger = logger.With(field.key, value)
		}
	}
	return logger, level
}

func LoadDotEnv() error {
	return godotenv.Load()
}

func ParseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Health(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "Health check requested")
	w.WriteHeader(http.StatusOK)
}

func Ready(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "Readiness check requested")
	w.WriteHeader(http.StatusOK)
}

func Run(serviceName, addr string, handler http.Handler) error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	shutdownTelemetry, err := observability.Setup(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("setup telemetry: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			slog.Error("failed to shutdown telemetry", "error", err)
		}
	}()

	server := newHTTPServer(serviceName, addr, handler)
	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("Starting server on", "addr", addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down HTTP server: %w", err)
	}

	err = <-serverErrors
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}

	slog.Info(serviceName + " service stopped")
	return nil
}

func newHTTPServer(serviceName, addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:           addr,
		Handler:        otelhttp.NewHandler(logRequests(handler), serviceName),
		ReadTimeout:    readHeaderTimeout,
		WriteTimeout:   shutdownTimeout,
		MaxHeaderBytes: 1 << 20,
	}
}
