// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package orders

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

type CreateItem struct {
	ProductID uuid.UUID
	Quantity  int
}

type CreateRequest struct {
	CustomerID uuid.UUID
	Items      []CreateItem
}

type CreateOrder struct {
	repository order.Repository
	newID      func() uuid.UUID
	now        func() time.Time
}

func NewCreateOrder(
	repository order.Repository,
	newID func() uuid.UUID,
	now func() time.Time,
) *CreateOrder {
	return &CreateOrder{
		repository: repository,
		newID:      newID,
		now:        now,
	}
}

func (c *CreateOrder) Execute(
	ctx context.Context,
	request CreateRequest,
) (*order.Order, error) {
	items := make([]order.Item, len(request.Items))

	for i, item := range request.Items {
		items[i] = order.Item{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	o, err := order.New(
		c.newID(),
		request.CustomerID,
		items,
		c.now(),
	)
	if err != nil {
		return nil, err
	}

	if err := c.repository.Save(ctx, o); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}

	return o, nil
}
