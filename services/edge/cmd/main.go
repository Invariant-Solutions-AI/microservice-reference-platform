// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/edge/internal/config"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/edge/internal/server"
	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/observability"
	"github.com/joho/godotenv"
)

const (
	serverName        = "edge"           // Name reported to the telemetry system.
	shutdownTimeout   = 10 * time.Second // Maximum time allowed for graceful shutdown.
	readHeaderTimeout = 5 * time.Second  // Maximum time allowed to read request headers.
)

func main() {
	// Keep main small: run contains the startup and shutdown lifecycle.
	if err := run(); err != nil {
		log.Fatalf("failed to run: %s", err)
		os.Exit(1)
	}
}

func run() error {
	// Cancel ctx when the process receives an interrupt (Ctrl+C) or SIGTERM.
	// This lets the service begin graceful shutdown instead of exiting immediately.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	defer stop()

	// Load optional environment variables from a local .env file.
	// Missing .env files are allowed because production environments usually
	// provide their configuration through the process environment.
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not loaded")
	}

	// Read the edge address and upstream Orders service URL.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
		os.Exit(1)
	}

	// Configure tracing before constructing or starting the HTTP server.
	// shutdownTelemetry flushes and closes the tracing provider.
	shutdownTelemetry, err := observability.Setup(
		ctx,
		serverName,
	)
	if err != nil {
		log.Fatalf("failed to setup telemetry: %s", err)
		os.Exit(1)
	}

	defer func() {
		// Give tracing a bounded amount of time to flush pending spans.
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			log.Fatalf("failed to shutdown telemetry: %s", err)
		}
	}()

	// Create the HTTP handler. It contains the edge routes and proxies orders
	// requests to the upstream URL supplied in the configuration.
	handler := server.NewHandler(cfg.OrdersURL)

	// Configure the HTTP server:
	// - Addr is the address to listen on.
	// - Handler receives every incoming request.
	// - ReadTimeout limits how long a slow client has to send its headers.
	// - WriteTimeout limits how long a response may take to be written.
	// - MaxHeaderBytes prevents an individual request from sending huge headers.
	s := &http.Server{
		Addr:           cfg.Addr,
		Handler:        handler,
		ReadTimeout:    readHeaderTimeout,
		WriteTimeout:   shutdownTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// A buffered channel carries the ListenAndServe result back to run.
	// Buffering allows the server goroutine to finish even if run is not
	// receiving from the channel at that exact moment.
	serverErrors := make(chan error, 1)

	// ListenAndServe blocks, so run it separately and report its result.
	go func() {
		log.Println("Starting server on", cfg.Addr)
		serverErrors <- s.ListenAndServe()
	}()

	// Continue in one of two cases:
	// - the HTTP server stops by itself or fails to start;
	// - the process receives a shutdown signal.
	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		// Stop signal forwarding before the explicit shutdown sequence.
		stop()
	}

	// Bound the graceful HTTP shutdown so it cannot wait forever.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf(
			"shut down HTTP server: %w",
			err,
		)
	}

	// ListenAndServe should now return http.ErrServerClosed. Any other error
	// means the server failed while it was stopping.
	err = <-serverErrors
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf(
			"serve HTTP during shutdown: %w",
			err,
		)
	}

	log.Println("edge service stopped")

	return nil
}
