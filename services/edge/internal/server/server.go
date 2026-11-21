// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import "net/http"

// create the HTTP handler for the edge service
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	// define your routes
	registerRoutes(mux)

	return mux
}