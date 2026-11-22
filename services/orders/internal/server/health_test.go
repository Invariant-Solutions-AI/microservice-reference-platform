// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHealthEndpoints(t *testing.T) {
	for _, tt := range []struct {
		name   string
		path   string
		ping   func(context.Context) error
		status int
	}{
		{name: "healthy without database", path: "/health", status: http.StatusOK},
		{name: "ready", path: "/ready", ping: func(context.Context) error { return nil }, status: http.StatusOK},
		{name: "database unavailable", path: "/ready", ping: func(context.Context) error { return errors.New("connection failed") }, status: http.StatusServiceUnavailable},
		{name: "database not configured", path: "/ready", status: http.StatusServiceUnavailable},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(nil, nil, tt.ping)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
		})
	}
}

func TestReadyUsesRequestContextAndTimeout(t *testing.T) {
	type contextKey struct{}
	key := contextKey{}
	handler := NewHandler(nil, nil, func(ctx context.Context) error {
		if ctx.Value(key) != "request" {
			t.Error("ping did not receive request context")
		}
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 2*time.Second {
			t.Error("ping did not receive the readiness deadline")
		}
		return ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), key, "request"))
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestReadyPostgres(t *testing.T) {
	databaseURL := os.Getenv("ORDERS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set ORDERS_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	pool, err := pgxpool.New(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)
	handler := NewHandler(nil, nil, pool.Ping)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	pool.Close()
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status after pool closed = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
