// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/persistence/memory"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/server"
)

func NewHandler() http.Handler {
	repository := memory.NewOrderRepository()
	createOrder := orders.NewCreateOrder(repository, uuid.New, time.Now)
	getOrder := orders.NewGetOrder(repository)
	return server.NewHandler(createOrder, getOrder)
}
