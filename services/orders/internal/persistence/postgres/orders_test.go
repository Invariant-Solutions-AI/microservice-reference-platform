// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

func testRepository(t *testing.T) (*OrderRepository, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("ORDERS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set ORDERS_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create admin pool: %v", err)
	}
	t.Cleanup(admin.Close)

	schema := "orders_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, err := admin.Exec(cleanupCtx, "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)

	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "000001_create_orders.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	up, _, found := strings.Cut(string(migration), "-- +goose Down")
	if !found {
		t.Fatal("migration is missing the down marker")
	}
	if _, err := pool.Exec(ctx, up); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	return NewOrderRepository(pool), pool
}

func testOrder(t *testing.T) *order.Order {
	t.Helper()

	o, err := order.New(
		uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
		[]order.Item{
			{ProductID: uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd"), Quantity: 3},
			{ProductID: uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc"), Quantity: 2},
		},
		time.Date(2026, time.November, 19, 12, 34, 56, 123456000, time.UTC),
	)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	return o
}

func TestOrderRepositoryRoundTrip(t *testing.T) {
	t.Parallel()
	repo, _ := testRepository(t)
	want := testOrder(t)
	want.Status = order.StatusCompleted

	if err := repo.Save(t.Context(), want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := repo.Get(t.Context(), want.OrderID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.OrderID != want.OrderID || got.CustomerID != want.CustomerID ||
		got.Status != want.Status || !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
	// Get orders items by product ID, not insertion order.
	wantItems := []order.Item{want.Items[1], want.Items[0]}
	if !reflect.DeepEqual(got.Items, wantItems) {
		t.Errorf("Get() items = %+v, want %+v", got.Items, wantItems)
	}
}

func TestOrderRepositoryDuplicate(t *testing.T) {
	t.Parallel()
	repo, _ := testRepository(t)
	original := testOrder(t)
	if err := repo.Save(t.Context(), original); err != nil {
		t.Fatalf("first Save() error = %v", err)
	}

	duplicate := *original
	duplicate.CustomerID = uuid.New()
	duplicate.Items = []order.Item{{ProductID: uuid.New(), Quantity: 9}}
	if err := repo.Save(t.Context(), &duplicate); !errors.Is(err, order.ErrOrderAlreadyExists) {
		t.Fatalf("duplicate Save() error = %v, want ErrOrderAlreadyExists", err)
	}

	got, err := repo.Get(t.Context(), original.OrderID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	wantItems := []order.Item{original.Items[1], original.Items[0]}
	if got.CustomerID != original.CustomerID || !reflect.DeepEqual(got.Items, wantItems) {
		t.Errorf("duplicate Save() changed stored order: %+v", got)
	}
}

func TestOrderRepositoryNotFound(t *testing.T) {
	t.Parallel()
	repo, _ := testRepository(t)
	got, err := repo.Get(t.Context(), uuid.New())
	if got != nil || !errors.Is(err, order.ErrOrderNotFound) {
		t.Errorf("Get() = %v, %v; want nil, ErrOrderNotFound", got, err)
	}
}

func TestOrderRepositorySaveRollsBackOnItemFailure(t *testing.T) {
	t.Parallel()
	repo, pool := testRepository(t)
	o := testOrder(t)
	o.Items[1].Quantity = 0

	if err := repo.Save(t.Context(), o); err == nil {
		t.Fatal("Save() succeeded with an invalid item")
	}
	if got, err := repo.Get(t.Context(), o.OrderID); got != nil || !errors.Is(err, order.ErrOrderNotFound) {
		t.Errorf("Get() after failed Save() = %v, %v; want nil, ErrOrderNotFound", got, err)
	}
	var itemCount int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM order_items WHERE order_id = $1", o.OrderID).Scan(&itemCount); err != nil {
		t.Fatalf("count items: %v", err)
	}
	if itemCount != 0 {
		t.Errorf("items after failed Save() = %d, want 0", itemCount)
	}

	o.Items[1].Quantity = 2
	if err := repo.Save(t.Context(), o); err != nil {
		t.Fatalf("Save() after rollback error = %v", err)
	}
}

func TestOrderRepositoryTestDataIsIsolated(t *testing.T) {
	t.Parallel()
	first, _ := testRepository(t)
	second, _ := testRepository(t)
	o := testOrder(t)

	if err := first.Save(t.Context(), o); err != nil {
		t.Fatalf("first Save() error = %v", err)
	}
	if got, err := second.Get(t.Context(), o.OrderID); got != nil || !errors.Is(err, order.ErrOrderNotFound) {
		t.Fatalf("second Get() = %v, %v; want nil, ErrOrderNotFound", got, err)
	}
	if err := second.Save(t.Context(), o); err != nil {
		t.Fatalf("second Save() error = %v", err)
	}
}
