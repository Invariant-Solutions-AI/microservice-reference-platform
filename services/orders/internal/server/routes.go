// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
)

func registerRoutes(mux *http.ServeMux, create *orders.CreateOrder) {
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/ready", ready)
	mux.HandleFunc("/orders", createOrder(create))
}
