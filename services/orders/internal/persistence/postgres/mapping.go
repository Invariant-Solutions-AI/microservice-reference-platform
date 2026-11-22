// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

type orderRow struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
	Status     string
	CreatedAt  time.Time
}

type orderItemRow struct {
	ProductID uuid.UUID
	Quantity  int
}

func toDomainOrder(
	row orderRow,
	itemRows []orderItemRow,
) (*order.Order, error) {
	items := make(
		[]order.Item,
		len(itemRows),
	)

	for i, itemRow := range itemRows {
		items[i] = order.Item{
			ProductID: itemRow.ProductID,
			Quantity:  itemRow.Quantity,
		}
	}

	status, err := parseStatus(row.Status)
	if err != nil {
		return nil, err
	}

	o, err := order.Rehydrate(
		row.OrderID,
		row.CustomerID,
		items,
		status,
		row.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"rehydrate order: %w",
			err,
		)
	}

	return o, nil
}

func parseStatus(
	value string,
) (order.Status, error) {
	status := order.Status(value)

	switch status {
	case order.StatusPending,
		order.StatusCompleted,
		order.StatusCancelled:
		return status, nil

	default:
		return "", fmt.Errorf(
			"%w: %q",
			order.ErrInvalidStatus,
			value,
		)
	}
}
