// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/persistence/memory"
)

func TestGetOrder(t *testing.T) {
	orderID := uuid.MustParse(
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	)
	customerID := uuid.MustParse(
		"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
	)
	productID := uuid.MustParse(
		"cccccccc-cccc-4ccc-8ccc-cccccccccccc",
	)

	createdAt := time.Date(
		2026,
		time.November,
		19,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	repository := memory.NewOrderRepository()

	existing, err := order.New(
		orderID,
		customerID,
		[]order.Item{
			{
				ProductID: productID,
				Quantity:  2,
			},
		},
		createdAt,
	)
	if err != nil {
		t.Fatalf("order.New() error = %v", err)
	}

	if err := repository.Save(t.Context(), existing); err != nil {
		t.Fatalf("repository.Save() error = %v", err)
	}

	createOrder := orders.NewCreateOrder(
		repository,
		uuid.New,
		time.Now,
	)

	getOrder := orders.NewGetOrder(repository)

	handler := NewHandler(
		createOrder,
		getOrder,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders/"+orderID.String(),
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			rec.Code,
			http.StatusOK,
			rec.Body.String(),
		)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf(
			"Content-Type = %q, want %q",
			contentType,
			"application/json",
		)
	}

	var got orderResponse

	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.OrderID != orderID {
		t.Errorf(
			"OrderID = %v, want %v",
			got.OrderID,
			orderID,
		)
	}

	if got.CustomerID != customerID {
		t.Errorf(
			"CustomerID = %v, want %v",
			got.CustomerID,
			customerID,
		)
	}

	if got.Status != string(order.StatusPending) {
		t.Errorf(
			"Status = %q, want %q",
			got.Status,
			order.StatusPending,
		)
	}

	if !got.CreatedAt.Equal(createdAt) {
		t.Errorf(
			"CreatedAt = %v, want %v",
			got.CreatedAt,
			createdAt,
		)
	}

	if len(got.Items) != 1 {
		t.Fatalf(
			"len(Items) = %d, want 1",
			len(got.Items),
		)
	}

	if got.Items[0].ProductID != productID {
		t.Errorf(
			"Items[0].ProductID = %v, want %v",
			got.Items[0].ProductID,
			productID,
		)
	}

	if got.Items[0].Quantity != 2 {
		t.Errorf(
			"Items[0].Quantity = %d, want 2",
			got.Items[0].Quantity,
		)
	}
}

func TestGetOrderReturnsBadRequestForInvalidOrderID(t *testing.T) {
	repository := memory.NewOrderRepository()

	createOrder := orders.NewCreateOrder(
		repository,
		uuid.New,
		time.Now,
	)

	getOrder := orders.NewGetOrder(repository)

	handler := NewHandler(
		createOrder,
		getOrder,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders/not-a-uuid",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			rec.Code,
			http.StatusBadRequest,
			rec.Body.String(),
		)
	}

	assertProblemDetails(
		t,
		rec,
		http.StatusBadRequest,
	)
}

func TestGetOrderReturnsNotFound(t *testing.T) {
	repository := memory.NewOrderRepository()

	createOrder := orders.NewCreateOrder(
		repository,
		uuid.New,
		time.Now,
	)

	getOrder := orders.NewGetOrder(repository)

	handler := NewHandler(
		createOrder,
		getOrder,
	)

	orderID := uuid.MustParse(
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders/"+orderID.String(),
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			rec.Code,
			http.StatusNotFound,
			rec.Body.String(),
		)
	}

	assertProblemDetails(
		t,
		rec,
		http.StatusNotFound,
	)
}

func assertProblemDetails(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	wantStatus int,
) {
	t.Helper()

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/problem+json" {
		t.Errorf(
			"Content-Type = %q, want %q",
			contentType,
			"application/problem+json",
		)
	}

	var problem problemDetails

	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf(
			"decode problem response: %v",
			err,
		)
	}

	if problem.Status != wantStatus {
		t.Errorf(
			"problem.Status = %d, want %d",
			problem.Status,
			wantStatus,
		)
	}

	if problem.Title != http.StatusText(wantStatus) {
		t.Errorf(
			"problem.Title = %q, want %q",
			problem.Title,
			http.StatusText(wantStatus),
		)
	}
}
