// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"log/slog"
	"testing"
)

func TestLoadServiceName(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "configured", value: "order-service", want: "order-service"},
		{name: "default", value: "", want: "orders"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ORDERS_SERVICE_NAME", tt.value)
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg.ServiceName != tt.want {
				t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, tt.want)
			}
		})
	}
}

func TestLoadDebugLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
}
