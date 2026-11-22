// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"
	"net/url"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/edge/internal/routing"
)

// create the HTTP handler for the edge service
func NewHandler(ordersURL *url.URL) http.Handler {
	mux := http.NewServeMux()

	proxy := routing.NewReverseProxy(ordersURL)

	orders := routing.Orders(proxy)

	// define your routes
	registerRoutes(mux, orders)

	return mux
}
