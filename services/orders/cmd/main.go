// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/config"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/persistence/memory"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/server"
	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/service"
)

func main() {
	service.LoadDotEnv()

	// Read the orders service address and upstream Orders service URL.
	cfg, err := config.Load()
	level := service.NewLogger(cfg.ServiceName)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	level.Set(cfg.LogLevel)

	repository := memory.NewOrderRepository()
	createOrder := orders.NewCreateOrder(repository, uuid.New, time.Now)
	getOrder := orders.NewGetOrder(repository)
	handler := server.NewHandler(createOrder, getOrder)

	if err := service.Run(cfg.ServiceName, cfg.Addr, handler); err != nil {
		slog.Error("failed to run the application", "error", err)
		os.Exit(1)
	}
}
