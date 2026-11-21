// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
)

func registerRoutes(mux *http.ServeMux, create *orders.CreateOrder, get *orders.GetOrder) {
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /ready", ready)
	mux.HandleFunc("POST /orders", createOrder(create))
	mux.HandleFunc("GET /orders/{orderId}", getOrder(get))
}
