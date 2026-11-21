// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

func TestOrderRepositorySaveAndGet(t *testing.T) {
	repo := NewOrderRepository()
	ctx := context.Background()

	createdAt := time.Date(
		2026,
		time.November,
		19,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	want, err := order.New(
		uuid.New(),
		uuid.New(),
		[]order.Item{
			{
				ProductID: uuid.New(),
				Quantity:  2,
			},
		},
		createdAt,
	)
	if err != nil {
		t.Fatalf("order.New() error = %v", err)
	}

	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := repo.Get(ctx, want.OrderID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestOrderRepositorySaveDuplicate(t *testing.T) {
	repo := NewOrderRepository()
	ctx := context.Background()

	o, err := order.New(
		uuid.New(),
		uuid.New(),
		[]order.Item{
			{
				ProductID: uuid.New(),
				Quantity:  1,
			},
		},
		time.Now(),
	)
	if err != nil {
		t.Fatalf("order.New() error = %v", err)
	}

	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("first Save() error = %v", err)
	}

	err = repo.Save(ctx, o)

	if !errors.Is(err, order.ErrOrderAlreadyExists) {
		t.Errorf(
			"second Save() error = %v, want %v",
			err,
			order.ErrOrderAlreadyExists,
		)
	}
}

func TestOrderRepositoryGetNotFound(t *testing.T) {
	repo := NewOrderRepository()

	_, err := repo.Get(context.Background(), uuid.New())

	if !errors.Is(err, order.ErrOrderNotFound) {
		t.Errorf(
			"Get() error = %v, want %v",
			err,
			order.ErrOrderNotFound,
		)
	}
}
