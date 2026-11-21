// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package orders

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

type fakeRepository struct {
	saved   *order.Order
	saveErr error
}

func (r *fakeRepository) Save(
	_ context.Context,
	o *order.Order,
) error {
	if r.saveErr != nil {
		return r.saveErr
	}

	r.saved = o

	return nil
}

func (r *fakeRepository) Get(
	_ context.Context,
	_ uuid.UUID,
) (*order.Order, error) {
	return nil, order.ErrOrderNotFound
}

func TestCreateOrderExecute(t *testing.T) {
	orderID := uuid.MustParse(
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	)
	customerID := uuid.MustParse(
		"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
	)
	productID := uuid.MustParse(
		"cccccccc-cccc-4ccc-8ccc-cccccccccccc",
	)

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

	repository := &fakeRepository{}

	createOrder := NewCreateOrder(
		repository,
		func() uuid.UUID {
			return orderID
		},
		func() time.Time {
			return createdAt
		},
	)

	got, err := createOrder.Execute(
		context.Background(),
		CreateRequest{
			CustomerID: customerID,
			Items: []CreateItem{
				{
					ProductID: productID,
					Quantity:  2,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got == nil {
		t.Fatal("Execute() order = nil, want order")
	}

	if got.OrderID != orderID {
		t.Errorf(
			"OrderID = %v, want %v",
			got.OrderID,
			orderID,
		)
	}

	if got.CustomerID != customerID {
		t.Errorf(
			"CustomerID = %v, want %v",
			got.CustomerID,
			customerID,
		)
	}

	if got.Status != order.StatusPending {
		t.Errorf(
			"Status = %v, want %v",
			got.Status,
			order.StatusPending,
		)
	}

	if !got.CreatedAt.Equal(createdAt) {
		t.Errorf(
			"CreatedAt = %v, want %v",
			got.CreatedAt,
			createdAt,
		)
	}

	if len(got.Items) != 1 {
		t.Fatalf(
			"len(Items) = %d, want 1",
			len(got.Items),
		)
	}

	if got.Items[0].ProductID != productID {
		t.Errorf(
			"Items[0].ProductID = %v, want %v",
			got.Items[0].ProductID,
			productID,
		)
	}

	if got.Items[0].Quantity != 2 {
		t.Errorf(
			"Items[0].Quantity = %d, want 2",
			got.Items[0].Quantity,
		)
	}

	if repository.saved == nil {
		t.Fatal("repository.Save() was not called")
	}

	if repository.saved != got {
		t.Error("repository.Save() did not receive the created order")
	}
}

func TestCreateOrderExecuteReturnsValidationError(t *testing.T) {
	repository := &fakeRepository{}

	createOrder := NewCreateOrder(
		repository,
		uuid.New,
		time.Now,
	)

	got, err := createOrder.Execute(
		context.Background(),
		CreateRequest{
			CustomerID: uuid.New(),
			Items:      nil,
		},
	)

	if !errors.Is(err, order.ErrItemsRequired) {
		t.Errorf(
			"Execute() error = %v, want %v",
			err,
			order.ErrItemsRequired,
		)
	}

	if got != nil {
		t.Errorf(
			"Execute() order = %+v, want nil",
			got,
		)
	}

	if repository.saved != nil {
		t.Error("repository.Save() was called for invalid order")
	}
}

func TestCreateOrderExecuteReturnsRepositoryError(t *testing.T) {
	errSave := errors.New("save failed")

	repository := &fakeRepository{
		saveErr: errSave,
	}

	createOrder := NewCreateOrder(
		repository,
		uuid.New,
		time.Now,
	)

	got, err := createOrder.Execute(
		context.Background(),
		CreateRequest{
			CustomerID: uuid.New(),
			Items: []CreateItem{
				{
					ProductID: uuid.New(),
					Quantity:  1,
				},
			},
		},
	)

	if !errors.Is(err, errSave) {
		t.Errorf(
			"Execute() error = %v, want wrapped %v",
			err,
			errSave,
		)
	}

	if got != nil {
		t.Errorf(
			"Execute() order = %+v, want nil",
			got,
		)
	}
}