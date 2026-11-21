// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import "net/http"

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/ready", ready)
}