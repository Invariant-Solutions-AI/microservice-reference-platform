// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"log/slog"
	"net"
	"net/url"
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
	DatabaseURL string
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

	host := envOrDefault("POSTGRES_HOST", "localhost")
	port := envOrDefault("POSTGRES_PORT", "5432")
	database := envOrDefault("POSTGRES_DATABASE", "orders")
	user := envOrDefault("POSTGRES_USER", "orders")
	params := url.Values{"sslmode": {envOrDefault("PGSSLMODE", "verify-full")}}
	databaseURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, os.Getenv("POSTGRES_PASSWORD")),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + database,
		RawPath:  "/" + url.PathEscape(database),
		RawQuery: params.Encode(),
	}

	return Config{
		ServiceName: serviceName,
		Addr:        addr,
		LogLevel:    logLevel,
		DatabaseURL: databaseURL.String(),
	}, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
