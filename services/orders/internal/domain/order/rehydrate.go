// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package order

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidStatus = errors.New("invalid order status")

// Rehydrate reconstructs an existing order from persistent state.
func Rehydrate(
	orderID uuid.UUID,
	customerID uuid.UUID,
	items []Item,
	status Status,
	createdAt time.Time,
) (*Order, error) {
	if orderID == uuid.Nil {
		return nil, ErrInvalidOrderID
	}

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

		// Existing orders may predate the lower creation limit.
		if item.Quantity <= 0 || item.Quantity > math.MaxInt32 {
			return nil, ErrInvalidQuantity
		}
	}

	switch status {
	case StatusPending,
		StatusCompleted,
		StatusCancelled:
	default:
		return nil, ErrInvalidStatus
	}

	if createdAt.IsZero() {
		return nil, errors.New("created time is required")
	}

	return &Order{
		OrderID:    orderID,
		CustomerID: customerID,
		Items:      append([]Item(nil), items...),
		Status:     status,
		CreatedAt:  createdAt,
	}, nil
}
