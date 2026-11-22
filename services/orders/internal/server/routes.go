// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/service"
)

func registerRoutes(mux *http.ServeMux, create *orders.CreateOrder, get *orders.GetOrder) {
	mux.HandleFunc("GET /health", service.Health)
	mux.HandleFunc("GET /ready", service.Ready)
	mux.HandleFunc("POST /orders", createOrder(create))
	mux.HandleFunc("GET /orders/{orderId}", getOrder(get))
}
