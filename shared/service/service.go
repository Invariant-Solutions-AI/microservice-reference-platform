// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/observability"
	"github.com/joho/godotenv"
)

const (
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func NewLogger(serviceName string) *slog.LevelVar {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: level},
		),
	).With("service", serviceName)
	slog.SetDefault(logger)
	return level
}

func LoadDotEnv() {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not loaded")
	}
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

func Health(w http.ResponseWriter, _ *http.Request) {
	slog.Debug("Health check requested")
	w.WriteHeader(http.StatusOK)
}

func Ready(w http.ResponseWriter, _ *http.Request) {
	slog.Debug("Readiness check requested")
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

	server := &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    readHeaderTimeout,
		WriteTimeout:   shutdownTimeout,
		MaxHeaderBytes: 1 << 20,
	}
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
