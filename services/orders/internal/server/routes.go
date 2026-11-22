// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"net/http"
	"time"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/service"
)

func registerRoutes(mux *http.ServeMux, create *orders.CreateOrder, get *orders.GetOrder, ping func(context.Context) error) {
	mux.HandleFunc("GET /health", service.Health)
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if ping == nil || ping(ctx) != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /orders", createOrder(create))
	mux.HandleFunc("GET /orders/{orderId}", getOrder(get))
}
