// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

func getOrder(
	get *orders.GetOrder,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("orderId"))
		if err != nil {
			err := fmt.Errorf(
				"%w: %v",
				order.ErrInvalidOrderID,
				err,
			)
			handleGetOrderError(w, err)
			return
		}

		order, err := get.Execute(r.Context(), id)
		if err != nil {
			handleGetOrderError(w, err)
			return
		}

		response := newOrderResponse(order)

		w.Header().Set(
			"Content-Type",
			"application/json",
		)
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func handleGetOrderError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, order.ErrInvalidOrderID):
		writeProblem(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(err, order.ErrOrderNotFound):
		writeProblem(
			w,
			http.StatusNotFound,
			err.Error(),
		)

	default:
		writeProblem(
			w,
			http.StatusInternalServerError,
			"Internal server error",
		)
		return
	}
}
