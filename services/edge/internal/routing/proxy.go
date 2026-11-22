// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package routing

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewProxy creates a reverse proxy to the specified target URL.
func NewReverseProxy(
	target *url.URL,
) http.Handler {
	slog.Debug("Creating reverse proxy for target", "target", target.String())
	return &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			slog.Debug("Rewriting request for proxy", "original_path", req.In.URL.Path, "target", target.String())
			req.SetURL(target)
			slog.Debug("Request URL set for proxy", "new_path", req.Out.URL.Path)
			req.SetXForwarded()
		},
		ErrorHandler: proxyErrorHandler,
	}
}

func proxyErrorHandler(
	w http.ResponseWriter,
	_ *http.Request,
	err error) {
	slog.Error("Proxy error", "error", err)
	http.Error(
		w,
		http.StatusText(http.StatusBadGateway),
		http.StatusBadGateway,
	)
}
