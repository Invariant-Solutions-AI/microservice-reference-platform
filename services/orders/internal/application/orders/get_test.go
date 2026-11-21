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

type getFakeRepository struct {
	orderID uuid.UUID
	result  *order.Order
	getErr  error
}

func (r *getFakeRepository) Save(
	_ context.Context,
	_ *order.Order,
) error {
	return nil
}

func (r *getFakeRepository) Get(
	_ context.Context,
	orderID uuid.UUID,
) (*order.Order, error) {
	r.orderID = orderID

	if r.getErr != nil {
		return nil, r.getErr
	}

	return r.result, nil
}

func TestGetOrderExecute(t *testing.T) {
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

	expected, err := order.New(
		orderID,
		customerID,
		[]order.Item{
			{
				ProductID: productID,
				Quantity:  2,
			},
		},
		createdAt,
	)
	if err != nil {
		t.Fatalf("order.New() error = %v", err)
	}

	repository := &getFakeRepository{
		result: expected,
	}

	getOrder := NewGetOrder(repository)

	got, err := getOrder.Execute(
		context.Background(),
		orderID,
	)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if repository.orderID != orderID {
		t.Errorf(
			"repository.Get() OrderID = %v, want %v",
			repository.orderID,
			orderID,
		)
	}

	if got != expected {
		t.Errorf(
			"Execute() order = %p, want %p",
			got,
			expected,
		)
	}
}

func TestGetOrderExecuteReturnsNotFound(t *testing.T) {
	orderID := uuid.MustParse(
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	)

	repository := &getFakeRepository{
		getErr: order.ErrOrderNotFound,
	}

	getOrder := NewGetOrder(repository)

	got, err := getOrder.Execute(
		context.Background(),
		orderID,
	)

	if !errors.Is(err, order.ErrOrderNotFound) {
		t.Errorf(
			"Execute() error = %v, want %v",
			err,
			order.ErrOrderNotFound,
		)
	}

	if got != nil {
		t.Errorf(
			"Execute() order = %+v, want nil",
			got,
		)
	}

	if repository.orderID != orderID {
		t.Errorf(
			"repository.Get() OrderID = %v, want %v",
			repository.orderID,
			orderID,
		)
	}
}

func TestGetOrderExecuteReturnsRepositoryError(t *testing.T) {
	orderID := uuid.MustParse(
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	)

	errRepository := errors.New("repository failure")

	repository := &getFakeRepository{
		getErr: errRepository,
	}

	getOrder := NewGetOrder(repository)

	got, err := getOrder.Execute(
		context.Background(),
		orderID,
	)

	if !errors.Is(err, errRepository) {
		t.Errorf(
			"Execute() error = %v, want %v",
			err,
			errRepository,
		)
	}

	if got != nil {
		t.Errorf(
			"Execute() order = %+v, want nil",
			got,
		)
	}
}
