// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/service"
)

const (
	defaultEdgePort    = ":8080"
	defaultServiceName = "edge"
)

type Config struct {
	ServiceName string
	Addr        string
	OrdersURL   *url.URL
	LogLevel    slog.Level
}

func Load() (Config, error) {
	serviceName := os.Getenv("EDGE_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	cfg := Config{ServiceName: serviceName}
	addr := os.Getenv("EDGE_ADDR")
	if addr == "" {
		addr = defaultEdgePort
	}
	slog.Debug("EDGE_ADDR", "value", addr)
	value := os.Getenv("ORDERS_URL")
	if value == "" {
		return cfg, fmt.Errorf(
			"ORDERS_URL is required",
		)
	}
	slog.Debug("ORDERS_URL", "value", value)

	ordersURL, err := url.Parse(value)
	if err != nil {
		return cfg, fmt.Errorf(
			"parse ORDERS_URL: %w",
			err,
		)
	}
	slog.Debug("Parsed ORDERS_URL", "value", ordersURL.String())
	if ordersURL.Scheme == "" ||
		ordersURL.Host == "" {
		return cfg, fmt.Errorf(
			"ORDERS_URL must be an absolute URL",
		)
	}

	logLevel := service.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	slog.Debug("LOG_LEVEL", "value", logLevel)

	return Config{
		ServiceName: serviceName,
		Addr:        addr,
		OrdersURL:   ordersURL,
		LogLevel:    logLevel,
	}, nil
}
