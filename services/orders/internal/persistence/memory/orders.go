// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"sync"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
	"github.com/google/uuid"
)

type OrderRepository struct {
	mux sync.Mutex
	orders map[string]*order.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[string]*order.Order),
	}
}

func (r *OrderRepository) Save(ctx context.Context, o *order.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	
	r.mux.Lock()
	defer r.mux.Unlock()

	if _, exists := r.orders[o.OrderID.String()]; exists {
		return order.ErrOrderAlreadyExists
	}
	
	r.orders[o.OrderID.String()] = cloneOrder(o)
	
	return nil
}

func (r *OrderRepository) Get(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	o, exists := r.orders[id.String()]
	if !exists {
		return nil, order.ErrOrderNotFound
	}

	return cloneOrder(o), nil
}

func cloneOrder(o *order.Order) *order.Order {
	if o == nil {
		return nil
	}

	clone := *o
	clone.Items = append([]order.Item(nil), o.Items...)

	return &clone
}