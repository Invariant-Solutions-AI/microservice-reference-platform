// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log/slog"
	"os"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/httpapi"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/config"
	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/service"
)

func main() {
	dotenvErr := service.LoadDotEnv()

	// Read the orders service address and upstream Orders service URL.
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

	handler := httpapi.NewHandler()

	if err := service.Run(cfg.ServiceName, cfg.Addr, handler); err != nil {
		slog.Error("failed to run the application", "error", err)
		os.Exit(1)
	}
}
