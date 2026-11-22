// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "health",
			path: "/health",
		},
		{
			name: "ready",
			path: "/ready",
		},
	}

	handler := NewHandler(nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				tt.path,
				nil,
			)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf(
					"status = %d, want %d",
					rec.Code,
					http.StatusOK,
				)
			}
		})
	}
}
