// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package routing

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewProxy creates a reverse proxy to the specified target URL.
func NewReverseProxy(
	target *url.URL,
) http.Handler {
	slog.Debug("Creating reverse proxy for target", "target", target.String())
	return &httputil.ReverseProxy{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Rewrite: func(req *httputil.ProxyRequest) {
			slog.DebugContext(req.In.Context(), "Rewriting request for proxy", "original_path", req.In.URL.Path, "target", target.String())
			req.SetURL(target)
			slog.DebugContext(req.Out.Context(), "Request URL set for proxy", "new_path", req.Out.URL.Path)
			req.SetXForwarded()
		},
		ErrorHandler: proxyErrorHandler,
	}
}

func proxyErrorHandler(
	w http.ResponseWriter,
	r *http.Request,
	err error) {
	slog.ErrorContext(r.Context(), "Proxy error", "error", err)
	http.Error(
		w,
		http.StatusText(http.StatusBadGateway),
		http.StatusBadGateway,
	)
}
