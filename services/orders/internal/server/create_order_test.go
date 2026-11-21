// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/application/orders"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/persistence/memory"
)

func TestCreateOrder(t *testing.T) {
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

	createOrder := orders.NewCreateOrder(
		repository,
		func() uuid.UUID {
			return orderID
		},
		func() time.Time {
			return createdAt
		},
	)

	getOrder := orders.NewGetOrder(repository)

	handler := NewHandler(createOrder, getOrder)

	body := []byte(`{
		"customerId": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		"items": [
			{
				"productId": "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
				"quantity": 2
			}
		]
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			rec.Code,
			http.StatusCreated,
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

	if got.Status != "pending" {
		t.Errorf(
			"Status = %q, want %q",
			got.Status,
			"pending",
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

	stored, err := repository.Get(
		context.Background(),
		orderID,
	)
	if err != nil {
		t.Fatalf("repository.Get() error = %v", err)
	}

	if stored.OrderID != orderID {
		t.Errorf(
			"stored OrderID = %v, want %v",
			stored.OrderID,
			orderID,
		)
	}

	if stored.CustomerID != customerID {
		t.Errorf(
			"stored CustomerID = %v, want %v",
			stored.CustomerID,
			customerID,
		)
	}

	if len(stored.Items) != 1 {
		t.Fatalf(
			"len(stored.Items) = %d, want 1",
			len(stored.Items),
		)
	}

	if stored.Items[0].ProductID != productID {
		t.Errorf(
			"stored ProductID = %v, want %v",
			stored.Items[0].ProductID,
			productID,
		)
	}
}

func TestCreateOrderReturnsBadRequest(t *testing.T) {
	customerID := uuid.MustParse(
		"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
	)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid JSON",
			body: `{`,
		},
		{
			name: "invalid customer ID",
			body: `{
				"customerId": "not-a-uuid",
				"items": [
					{
						"productId": "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
						"quantity": 1
					}
				]
			}`,
		},
		{
			name: "invalid product ID",
			body: `{
				"customerId": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
				"items": [
					{
						"productId": "not-a-uuid",
						"quantity": 1
					}
				]
			}`,
		},
		{
			name: "missing items",
			body: `{
				"customerId": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
			}`,
		},
		{
			name: "zero quantity",
			body: `{
				"customerId": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
				"items": [
					{
						"productId": "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
						"quantity": 0
					}
				]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := memory.NewOrderRepository()

			createOrder := orders.NewCreateOrder(
				repository,
				uuid.New,
				time.Now,
			)

			getOrder := orders.NewGetOrder(repository)

			handler := NewHandler(createOrder, getOrder)

			req := httptest.NewRequest(
				http.MethodPost,
				"/orders",
				bytes.NewBufferString(tt.body),
			)
			req.Header.Set(
				"Content-Type",
				"application/json",
			)

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"status = %d, want %d; body = %s",
					rec.Code,
					http.StatusBadRequest,
					rec.Body.String(),
				)
			}

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

			if problem.Status != http.StatusBadRequest {
				t.Errorf(
					"problem.Status = %d, want %d",
					problem.Status,
					http.StatusBadRequest,
				)
			}

			if problem.Title != http.StatusText(http.StatusBadRequest) {
				t.Errorf(
					"problem.Title = %q, want %q",
					problem.Title,
					http.StatusText(http.StatusBadRequest),
				)
			}
		})
	}

	_ = customerID
}
