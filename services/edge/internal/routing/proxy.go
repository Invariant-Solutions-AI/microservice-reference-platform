// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package routing

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewProxy creates a reverse proxy to the specified target URL.
func NewReverseProxy(
	target *url.URL,
) http.Handler {
	return &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.SetURL(target)
			req.SetXForwarded()
		},
		ErrorHandler: proxyErrorHandler,
	}
}

func proxyErrorHandler(
	w http.ResponseWriter,
	_ *http.Request,
	_ error) {
	http.Error(
		w,
		http.StatusText(http.StatusBadGateway),
		http.StatusBadGateway,
	)
}
