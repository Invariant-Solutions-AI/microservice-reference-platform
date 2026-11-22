// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

func createOrder(
	create *orders.CreateOrder,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request createOrderRequest

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			writeProblem(
				w,
				http.StatusBadRequest,
				"invalid request body",
			)
			return
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			writeProblem(
				w,
				http.StatusBadRequest,
				"invalid request body",
			)
			return
		}

		items := make(
			[]orders.CreateItem,
			len(request.Items),
		)

		for i, item := range request.Items {
			items[i] = orders.CreateItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			}
		}

		created, err := create.Execute(
			r.Context(),
			orders.CreateRequest{
				CustomerID: request.CustomerID,
				Items:      items,
			},
		)
		if err != nil {
			handleCreateOrderError(w, err)
			return
		}

		response := newOrderResponse(created)

		w.Header().Set(
			"Content-Type",
			"application/json",
		)
		w.WriteHeader(http.StatusCreated)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func handleCreateOrderError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, order.ErrCustomerIDRequired):
		writeProblem(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(err, order.ErrItemsRequired):
		writeProblem(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(err, order.ErrProductIDRequired):
		writeProblem(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(err, order.ErrInvalidQuantity):
		writeProblem(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(err, order.ErrOrderAlreadyExists):
		writeProblem(
			w,
			http.StatusConflict,
			err.Error(),
		)

	default:
		writeProblem(
			w,
			http.StatusInternalServerError,
			"unable to create order",
		)
	}
}
