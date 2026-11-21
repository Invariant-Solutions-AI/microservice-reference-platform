// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
)

// create the HTTP handler for the edge service
func NewHandler(create *orders.CreateOrder) http.Handler {
	mux := http.NewServeMux()
	// define your routes
	registerRoutes(mux, create)

	return mux
}
