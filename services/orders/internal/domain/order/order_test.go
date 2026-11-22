// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package order

import (
	"errors"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNew(t *testing.T) {
	createdAt := time.Date(2026, time.November, 19, 12, 0, 0, 0, time.UTC)

	items := []Item{
		{
			ProductID: uuid.New(),
			Quantity:  2,
		},
	}

	orderID := uuid.New()
	customerID := uuid.New()
	got, err := New(orderID, customerID, items, createdAt)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	if got.OrderID != orderID {
		t.Errorf("OrderID = %q, want %q", got.OrderID, orderID)
	}

	if got.CustomerID != customerID {
		t.Errorf("CustomerID = %q, want %q", got.CustomerID.String(), customerID.String())
	}

	if got.Status != StatusPending {
		t.Errorf("Status = %q, want %q", got.Status, StatusPending)
	}

	if !got.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, createdAt)
	}

	if len(got.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(got.Items))
	}

	if got.Items[0] != items[0] {
		t.Errorf("Items[0] = %+v, want %+v", got.Items[0], items[0])
	}
}

func TestQuantityRange(t *testing.T) {
	tooLarge := int64(math.MaxInt32) + 1
	tests := []struct {
		name          string
		quantity      int64
		wantErr       error
		wantRehydrate error
	}{
		{name: "maximum", quantity: 1000},
		{name: "above maximum", quantity: 1001, wantErr: ErrInvalidQuantity},
		{name: "previous maximum", quantity: math.MaxInt32, wantErr: ErrInvalidQuantity},
		{name: "above database range", quantity: tooLarge, wantErr: ErrInvalidQuantity, wantRehydrate: ErrInvalidQuantity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strconv.IntSize == 32 && tt.quantity > math.MaxInt32 {
				t.Skip("quantity cannot be represented as int on this platform")
			}
			items := []Item{{ProductID: uuid.New(), Quantity: int(tt.quantity)}}
			id, customerID, now := uuid.New(), uuid.New(), time.Now()
			_, err := New(id, customerID, items, now)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("New() error = %v, want %v", err, tt.wantErr)
			}
			_, err = Rehydrate(id, customerID, items, StatusPending, now)
			if !errors.Is(err, tt.wantRehydrate) {
				t.Errorf("Rehydrate() error = %v, want %v", err, tt.wantRehydrate)
			}
		})
	}
}

func TestNewValidation(t *testing.T) {
	createdAt := time.Date(2026, time.November, 19, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		customerID uuid.UUID
		items      []Item
		wantErr    error
	}{
		{
			name:       "customer ID is empty",
			customerID: uuid.Nil,
			items: []Item{
				{ProductID: uuid.New(), Quantity: 1},
			},
			wantErr: ErrCustomerIDRequired,
		},
		{
			name:       "items are nil",
			customerID: uuid.New(),
			items:      nil,
			wantErr:    ErrItemsRequired,
		},
		{
			name:       "items are empty",
			customerID: uuid.New(),
			items:      []Item{},
			wantErr:    ErrItemsRequired,
		},
		{
			name:       "product ID is empty",
			customerID: uuid.New(),
			items: []Item{
				{ProductID: uuid.Nil, Quantity: 1},
			},
			wantErr: ErrProductIDRequired,
		},
		{
			name:       "product ID is whitespace",
			customerID: uuid.New(),
			items: []Item{
				{ProductID: uuid.Nil, Quantity: 1},
			},
			wantErr: ErrProductIDRequired,
		},
		{
			name:       "quantity is zero",
			customerID: uuid.New(),
			items: []Item{
				{ProductID: uuid.New(), Quantity: 0},
			},
			wantErr: ErrInvalidQuantity,
		},
		{
			name:       "quantity is negative",
			customerID: uuid.New(),
			items: []Item{
				{ProductID: uuid.New(), Quantity: -1},
			},
			wantErr: ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(
				uuid.New(),
				tt.customerID,
				tt.items,
				createdAt,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("New() error = %v, want %v", err, tt.wantErr)
			}

			if got != nil {
				t.Errorf("New() order = %+v, want nil", got)
			}
		})
	}
}

func TestNewCopiesItems(t *testing.T) {
	items := []Item{
		{
			ProductID: uuid.New(),
			Quantity:  1,
		},
	}

	orderID := uuid.New()
	customerID := uuid.New()
	createdAt := time.Now()

	got, err := New(
		orderID,
		customerID,
		items,
		createdAt,
	)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	items[0].Quantity = 99

	if got.Items[0].Quantity != 1 {
		t.Errorf(
			"Items[0].Quantity = %d after source mutation, want 1",
			got.Items[0].Quantity,
		)
	}
}
