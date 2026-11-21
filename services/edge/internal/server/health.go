// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import "net/http"

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}