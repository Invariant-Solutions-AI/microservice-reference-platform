// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"log/slog"
	"os"

	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/service"
)

const (
	defaultOrdersPort  = ":8081"
	defaultServiceName = "orders"
)

type Config struct {
	ServiceName string
	Addr        string
	LogLevel    slog.Level
}

func Load() (Config, error) {
	serviceName := os.Getenv("ORDERS_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	addr := os.Getenv("ORDERS_ADDR")
	if addr == "" {
		addr = defaultOrdersPort
	}
	slog.Debug("ORDERS_ADDR", "value", addr)

	logLevel := service.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	slog.Debug("LOG_LEVEL", "value", logLevel)

	return Config{
		ServiceName: serviceName,
		Addr:        addr,
		LogLevel:    logLevel,
	}, nil
}
