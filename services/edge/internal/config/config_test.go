// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"log/slog"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("EDGE_SERVICE_NAME", "public-edge")
	t.Setenv("EDGE_ADDR", ":9090")
	t.Setenv("ORDERS_URL", "http://localhost:8081")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ServiceName != "public-edge" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "public-edge")
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}

	if cfg.Addr != ":9090" {
		t.Errorf(
			"Addr = %q, want %q",
			cfg.Addr,
			":9090",
		)
	}

	if got := cfg.OrdersURL.String(); got != "http://localhost:8081" {
		t.Errorf(
			"OrdersURL = %q, want %q",
			got,
			"http://localhost:8081",
		)
	}
}

func TestLoadUsesDefaultAddress(t *testing.T) {
	t.Setenv("EDGE_SERVICE_NAME", "")
	t.Setenv("EDGE_ADDR", "")
	t.Setenv("ORDERS_URL", "http://localhost:8081")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ServiceName != "edge" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "edge")
	}

	if cfg.Addr != ":8080" {
		t.Errorf(
			"Addr = %q, want %q",
			cfg.Addr,
			":8080",
		)
	}
}

func TestLoadRequiresOrdersURL(t *testing.T) {
	t.Setenv("EDGE_SERVICE_NAME", "public-edge")
	t.Setenv("ORDERS_URL", "")

	cfg, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if cfg.ServiceName != "public-edge" {
		t.Errorf("ServiceName on config error = %q, want %q", cfg.ServiceName, "public-edge")
	}
}

func TestLoadRejectsInvalidOrdersURL(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "missing scheme",
			value: "localhost:8081",
		},
		{
			name:  "missing host",
			value: "http://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(
				"ORDERS_URL",
				tt.value,
			)

			_, err := Load()
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}
		})
	}
}
