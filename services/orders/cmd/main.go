// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/persistence/memory"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/server"
	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/observability"
	"github.com/joho/godotenv"
)

const (
	serviceName       = "orders"
	defaultAddr       = ":8081"
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %s", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// read configuration
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not loaded")
	}

	addr := os.Getenv("ORDERS_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	shutdownTelemetry, err := observability.Setup(
		ctx,
		serviceName,
	)
	if err != nil {
		log.Printf("Error setting up telemetry: %s", err)
		return err
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := shutdownTelemetry(shutdownCtx); err != nil {
			log.Printf("Error shutting down telemetry: %s", err)
		}
	}()

	// Infrastructure
	repository := memory.NewOrderRepository()

	// Application
	createOrder := orders.NewCreateOrder(
		repository,
		uuid.New,
		time.Now,
	)

	getOrder := orders.NewGetOrder(
		repository,
	)

	// Transport
	handler := server.NewHandler(createOrder, getOrder)

	s := &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    readHeaderTimeout,
		WriteTimeout:   shutdownTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	serverErrors := make(chan error, 1)

	// start order service
	go func() {
		serverErrors <- s.ListenAndServe()
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

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down server: %s", err)
	}

	err = <-serverErrors
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
