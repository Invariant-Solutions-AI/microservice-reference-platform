// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/config"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/persistence/postgres"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/server"
	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to run the application", "error", err)
		os.Exit(1)
	}
}

func run() error {
	dotenvErr := service.LoadDotEnv()

	cfg, err := config.Load()
	level := service.NewLogger(cfg.ServiceName)
	if dotenvErr != nil {
		slog.Warn(".env file not loaded")
	}
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	level.Set(cfg.LogLevel)

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return errors.New("failed to configure PostgreSQL pool")
	}
	defer pool.Close()

	return service.Run(cfg.ServiceName, cfg.Addr, newHandler(pool))
}

func newHandler(pool *pgxpool.Pool) http.Handler {
	repository := postgres.NewOrderRepository(pool)
	createOrder := orders.NewCreateOrder(repository, uuid.New, time.Now)
	getOrder := orders.NewGetOrder(repository)
	return server.NewHandler(createOrder, getOrder, pool.Ping)
}
