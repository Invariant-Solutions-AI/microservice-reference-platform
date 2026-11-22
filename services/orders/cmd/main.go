// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/httpapi"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/config"
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

	handler := httpapi.NewHandler(pool.Ping)

	return service.Run(cfg.ServiceName, cfg.Addr, handler)
}
