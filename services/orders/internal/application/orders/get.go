// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package orders

import (
	"context"

	"github.com/google/uuid"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

type GetOrder struct {
	repository order.Repository
}

func NewGetOrder(repository order.Repository) *GetOrder {
	return &GetOrder{
		repository: repository,
	}
}

func (g *GetOrder) Execute(
	ctx context.Context,
	orderID uuid.UUID,
) (*order.Order, error) {
	return g.repository.Get(ctx, orderID)
}
