// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("EDGE_ADDR", ":9090")
	t.Setenv("ORDERS_URL", "http://localhost:8081")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
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
	t.Setenv("EDGE_ADDR", "")
	t.Setenv("ORDERS_URL", "http://localhost:8081")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
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
	t.Setenv("ORDERS_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
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
