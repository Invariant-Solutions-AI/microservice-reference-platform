// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"net/http"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
)

// create the HTTP handler for the edge service
func NewHandler(create *orders.CreateOrder, get *orders.GetOrder, ping func(context.Context) error) http.Handler {
	mux := http.NewServeMux()
	// define your routes
	registerRoutes(mux, create, get, ping)

	return mux
}
