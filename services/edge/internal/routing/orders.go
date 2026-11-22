// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package routing

import (
	"log/slog"
	"net/http"
	"strings"
)

func Orders(
	proxy http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Original request path", "path", r.URL.Path)
		path := strings.TrimPrefix(r.URL.Path, "/api/v1")
		slog.Debug("Trimmed request path", "path", path)

		req := r.Clone(r.Context())
		req.URL.Path = path
		slog.Debug("Request path set for proxy", "path", req.URL.Path)

		proxy.ServeHTTP(w, req)
	})
}
