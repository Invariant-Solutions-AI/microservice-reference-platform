// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"net/http"
)

// problemDetails represents an RFC 9457 Problem Details response.
type problemDetails struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func writeProblem(
	w http.ResponseWriter,
	status int,
	detail string,
) {
	problem := problemDetails{
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	}

	w.Header().Set(
		"Content-Type",
		"application/problem+json",
	)
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(problem)
}
