// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
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

func TestLoadPostgresSettings(t *testing.T) {
	for _, tt := range []struct {
		name     string
		host     string
		port     string
		database string
		user     string
		password string
		wantHost string
		wantPort uint16
		wantDB   string
		wantUser string
	}{
		{name: "defaults", wantHost: "localhost", wantPort: 5432, wantDB: "orders", wantUser: "orders"},
		{name: "configured", host: "::1", port: "5433", database: "orders/test", user: "orders@user", password: "p@ss:/word", wantHost: "::1", wantPort: 5433, wantDB: "orders/test", wantUser: "orders@user"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("POSTGRES_HOST", tt.host)
			t.Setenv("POSTGRES_PORT", tt.port)
			t.Setenv("POSTGRES_DATABASE", tt.database)
			t.Setenv("POSTGRES_USER", tt.user)
			t.Setenv("POSTGRES_PASSWORD", tt.password)
			t.Setenv("PGSSLMODE", "")
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
			if err != nil {
				t.Fatal("invalid database configuration")
			}
			conn := poolConfig.ConnConfig
			if conn.Host != tt.wantHost || conn.Port != tt.wantPort || conn.Database != tt.wantDB || conn.User != tt.wantUser || conn.Password != tt.password {
				t.Error("PostgreSQL connection settings do not match environment")
			}
			if conn.TLSConfig == nil || conn.TLSConfig.InsecureSkipVerify || conn.TLSConfig.ServerName != tt.wantHost || len(conn.Fallbacks) != 0 {
				t.Error("default PostgreSQL connection must verify the server without a plaintext fallback")
			}
		})
	}
}

func TestLoadPostgresTLSDisabledForLocalDevelopment(t *testing.T) {
	t.Setenv("PGSSLMODE", "disable")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if poolConfig.ConnConfig.TLSConfig != nil || len(poolConfig.ConnConfig.Fallbacks) != 0 {
		t.Fatal("expected PostgreSQL TLS to be disabled")
	}
}
