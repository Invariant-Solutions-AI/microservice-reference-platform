// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

func newOrderResponse(
	o *order.Order,
) orderResponse {
	items := make(
		[]orderItemResponse,
		len(o.Items),
	)

	for i, item := range o.Items {
		items[i] = orderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return orderResponse{
		OrderID:    o.OrderID,
		CustomerID: o.CustomerID,
		Items:      items,
		Status:     string(o.Status),
		CreatedAt:  o.CreatedAt,
	}
}
