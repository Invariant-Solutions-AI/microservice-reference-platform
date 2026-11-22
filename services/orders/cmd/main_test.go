// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlerPersistsOrdersAcrossInstances(t *testing.T) {
	databaseURL := os.Getenv("ORDERS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set ORDERS_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	schema := pgx.Identifier{"orders_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop schema: %v", err)
		}
	}()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	migration, err := os.ReadFile(filepath.Join("..", "migrations", "000001_create_orders.sql"))
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(migration), "-- +goose Down")
	if !found {
		t.Fatal("migration is missing the down marker")
	}
	if _, err := pool.Exec(ctx, up); err != nil {
		t.Fatal(err)
	}

	body := `{"customerId":"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb","items":[{"productId":"cccccccc-cccc-4ccc-8ccc-cccccccccccc","quantity":2}]}`
	created := httptest.NewRecorder()
	newHandler(pool).ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body)))
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", created.Code, created.Body)
	}
	var response struct {
		OrderID string `json:"orderId"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &response); err != nil || response.OrderID == "" {
		t.Fatalf("invalid create response: %s (%v)", created.Body, err)
	}

	retrieved := httptest.NewRecorder()
	newHandler(pool).ServeHTTP(retrieved, httptest.NewRequest(http.MethodGet, "/orders/"+response.OrderID, nil))
	if retrieved.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", retrieved.Code, retrieved.Body)
	}
	var got struct {
		OrderID    string `json:"orderId"`
		CustomerID string `json:"customerId"`
		Items      []struct {
			ProductID string `json:"productId"`
			Quantity  int    `json:"quantity"`
		} `json:"items"`
	}
	if err := json.Unmarshal(retrieved.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.OrderID != response.OrderID || got.CustomerID != "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb" ||
		len(got.Items) != 1 || got.Items[0].ProductID != "cccccccc-cccc-4ccc-8ccc-cccccccccccc" || got.Items[0].Quantity != 2 {
		t.Fatalf("unexpected retrieved order: %+v", got)
	}
}
