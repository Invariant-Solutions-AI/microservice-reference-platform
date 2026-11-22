// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package order

import (
	"time"

	"github.com/google/uuid"
)

const maxItemQuantity = 1000

type Item struct {
	ProductID uuid.UUID
	Quantity  int
}

type Order struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
	Items      []Item
	Status     Status
	CreatedAt  time.Time
}

func New(id uuid.UUID, customerID uuid.UUID, items []Item, CreatedAt time.Time) (*Order, error) {
	if customerID == uuid.Nil {
		return nil, ErrCustomerIDRequired
	}
	if len(items) == 0 {
		return nil, ErrItemsRequired
	}
	for _, item := range items {
		if item.ProductID == uuid.Nil {
			return nil, ErrProductIDRequired
		}
		if item.Quantity <= 0 || item.Quantity > maxItemQuantity {
			return nil, ErrInvalidQuantity
		}
	}
	return &Order{
		OrderID:    id,
		CustomerID: customerID,
		Items:      append([]Item(nil), items...),
		Status:     StatusPending,
		CreatedAt:  CreatedAt,
	}, nil
}
