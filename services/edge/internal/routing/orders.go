// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package routing

import (
	"net/http"
	"strings"
)

func Orders(
	proxy http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1")

		req := r.Clone(r.Context())
		req.URL.Path = path

		proxy.ServeHTTP(w, req)
	})
}
