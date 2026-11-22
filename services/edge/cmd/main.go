// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log/slog"
	"os"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/edge/internal/config"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/edge/internal/server"
	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/service"
)

func main() {
	dotenvErr := service.LoadDotEnv()

	// Read the edge address and upstream Orders service URL.
	cfg, err := config.Load()
	level := service.NewLogger(cfg.ServiceName)
	if dotenvErr != nil {
		slog.Warn(".env file not loaded")
	}
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	level.Set(cfg.LogLevel)

	handler := server.NewHandler(cfg.OrdersURL)
	if err := service.Run(cfg.ServiceName, cfg.Addr, handler); err != nil {
		slog.Error("failed to run the application", "error", err)
		os.Exit(1)
	}
}
