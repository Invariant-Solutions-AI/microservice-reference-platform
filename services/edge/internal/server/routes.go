// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"

	"github.com/invariant-solutions-ai/microservice-reference-platform/shared/service"
)

func registerRoutes(
	mux *http.ServeMux,
	orders http.Handler,
) {
	mux.HandleFunc("GET /health", service.Health)
	mux.HandleFunc("GET /ready", service.Ready)

	mux.Handle(
		"POST /api/v1/orders",
		orders,
	)

	mux.Handle(
		"GET /api/v1/orders/{orderId}",
		orders,
	)
}
