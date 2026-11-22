// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrdersRemovesAPIPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
	}{
		{
			name:     "orders collection",
			input:    "/api/v1/orders",
			wantPath: "/orders",
		},
		{
			name: "specific order",
			input: "/api/v1/orders/" +
				"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			wantPath: "/orders/" +
				"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string

			next := http.HandlerFunc(
				func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					gotPath = r.URL.Path

					w.WriteHeader(http.StatusOK)
				},
			)

			handler := Orders(next)

			req := httptest.NewRequest(
				http.MethodGet,
				tt.input,
				nil,
			)

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if gotPath != tt.wantPath {
				t.Errorf(
					"path = %q, want %q",
					gotPath,
					tt.wantPath,
				)
			}
		})
	}
}

func TestOrdersDoesNotMutateOriginalRequest(
	t *testing.T,
) {
	next := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		},
	)

	handler := Orders(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/orders",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if req.URL.Path != "/api/v1/orders" {
		t.Errorf(
			"original path = %q, want %q",
			req.URL.Path,
			"/api/v1/orders",
		)
	}
}
