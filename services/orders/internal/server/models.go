// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"time"

	"github.com/google/uuid"
)

type createOrderRequest struct {
	CustomerID uuid.UUID                `json:"customerId"`
	Items      []createOrderItemRequest `json:"items"`
}

type createOrderItemRequest struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
}

type orderResponse struct {
	OrderID    uuid.UUID           `json:"orderId"`
	CustomerID uuid.UUID           `json:"customerId"`
	Items      []orderItemResponse `json:"items"`
	Status     string              `json:"status"`
	CreatedAt  time.Time           `json:"createdAt"`
}

type orderItemResponse struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
}