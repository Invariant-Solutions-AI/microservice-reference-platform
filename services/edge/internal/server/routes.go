// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import "net/http"

func registerRoutes(
	mux *http.ServeMux,
	orders http.Handler,
) {
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /ready", ready)

	mux.Handle(
		"POST /api/v1/orders",
		orders,
	)

	mux.Handle(
		"GET /api/v1/orders/{orderId}",
		orders,
	)
}
