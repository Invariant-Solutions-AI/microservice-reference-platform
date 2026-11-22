// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	ordershttpapi "github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/httpapi"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

func TestEdgeToOrdersCreateAndRetrievePropagateTraceContext(t *testing.T) {
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()

	edgeExporter := tracetest.NewInMemoryExporter()
	edgeProvider := newTestTracerProvider(t, edgeExporter, "edge")
	ordersExporter := tracetest.NewInMemoryExporter()
	ordersProvider := newTestTracerProvider(t, ordersExporter, "orders")
	otel.SetTracerProvider(edgeProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
		_ = edgeProvider.Shutdown(t.Context())
		_ = ordersProvider.Shutdown(t.Context())
	})

	ordersServer := httptest.NewServer(otelhttp.NewHandler(
		ordershttpapi.NewHandler(nil),
		"orders",
		otelhttp.WithTracerProvider(ordersProvider),
	))
	t.Cleanup(ordersServer.Close)

	ordersURL, err := url.Parse(ordersServer.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	edgeServer := httptest.NewServer(otelhttp.NewHandler(
		NewHandler(ordersURL),
		"edge",
		otelhttp.WithTracerProvider(edgeProvider),
	))
	t.Cleanup(edgeServer.Close)

	customerID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	productID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	createTraceID := "11111111111111111111111111111111"
	createParentID := "2222222222222222"
	createBody := []byte(fmt.Sprintf(`{
		"customerId": %q,
		"items": [{"productId": %q, "quantity": 2}]
	}`, customerID, productID))

	createRequest, err := http.NewRequest(
		http.MethodPost,
		edgeServer.URL+"/api/v1/orders",
		bytes.NewReader(createBody),
	)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.Header.Set("traceparent", "00-"+createTraceID+"-"+createParentID+"-01")

	createResponse, err := edgeServer.Client().Do(createRequest)
	if err != nil {
		t.Fatalf("create order request error = %v", err)
	}
	defer createResponse.Body.Close()
	if createResponse.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResponse.StatusCode, http.StatusCreated)
	}

	var created integrationOrderResponse
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.OrderID == uuid.Nil {
		t.Fatal("create response orderId is nil")
	}
	if created.CustomerID != customerID {
		t.Errorf("create customerId = %s, want %s", created.CustomerID, customerID)
	}
	if len(created.Items) != 1 {
		t.Fatalf("create items = %d, want 1", len(created.Items))
	}
	if created.Items[0].ProductID != productID {
		t.Errorf("create productId = %s, want %s", created.Items[0].ProductID, productID)
	}
	if created.Items[0].Quantity != 2 {
		t.Errorf("create quantity = %d, want 2", created.Items[0].Quantity)
	}

	retrieveTraceID := "33333333333333333333333333333333"
	retrieveParentID := "4444444444444444"
	retrieveRequest, err := http.NewRequest(
		http.MethodGet,
		edgeServer.URL+"/api/v1/orders/"+created.OrderID.String(),
		nil,
	)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	retrieveRequest.Header.Set("traceparent", "00-"+retrieveTraceID+"-"+retrieveParentID+"-01")

	retrieveResponse, err := edgeServer.Client().Do(retrieveRequest)
	if err != nil {
		t.Fatalf("retrieve order request error = %v", err)
	}
	defer retrieveResponse.Body.Close()
	if retrieveResponse.StatusCode != http.StatusOK {
		t.Fatalf("retrieve status = %d, want %d", retrieveResponse.StatusCode, http.StatusOK)
	}

	var retrieved integrationOrderResponse
	if err := json.NewDecoder(retrieveResponse.Body).Decode(&retrieved); err != nil {
		t.Fatalf("decode retrieve response: %v", err)
	}
	if retrieved.OrderID != created.OrderID {
		t.Errorf("retrieve orderId = %s, want %s", retrieved.OrderID, created.OrderID)
	}
	if retrieved.CustomerID != customerID {
		t.Errorf("retrieve customerId = %s, want %s", retrieved.CustomerID, customerID)
	}
	if len(retrieved.Items) != 1 {
		t.Fatalf("retrieve items = %d, want 1", len(retrieved.Items))
	}
	if retrieved.Items[0].ProductID != productID {
		t.Errorf("retrieve productId = %s, want %s", retrieved.Items[0].ProductID, productID)
	}
	if retrieved.Items[0].Quantity != 2 {
		t.Errorf("retrieve quantity = %d, want 2", retrieved.Items[0].Quantity)
	}

	assertTraceFlow(t, edgeExporter.GetSpans(), ordersExporter.GetSpans(), createTraceID, createParentID)
	assertTraceFlow(t, edgeExporter.GetSpans(), ordersExporter.GetSpans(), retrieveTraceID, retrieveParentID)
}

type integrationOrderResponse struct {
	OrderID    uuid.UUID `json:"orderId"`
	CustomerID uuid.UUID `json:"customerId"`
	Items      []struct {
		ProductID uuid.UUID `json:"productId"`
		Quantity  int       `json:"quantity"`
	} `json:"items"`
}

func assertTraceFlow(
	t *testing.T,
	edgeSpans []tracetest.SpanStub,
	ordersSpans []tracetest.SpanStub,
	traceIDHex string,
	rootParentIDHex string,
) {
	t.Helper()

	wantTraceID, err := trace.TraceIDFromHex(traceIDHex)
	if err != nil {
		t.Fatalf("trace.TraceIDFromHex() error = %v", err)
	}
	wantRootParentID, err := trace.SpanIDFromHex(rootParentIDHex)
	if err != nil {
		t.Fatalf("trace.SpanIDFromHex() error = %v", err)
	}

	var edgeServerSpan, edgeClientSpan *tracetest.SpanStub
	edgeTraceSpanCount := 0
	for i := range edgeSpans {
		span := &edgeSpans[i]
		if span.SpanContext.TraceID() != wantTraceID {
			continue
		}
		edgeTraceSpanCount++
		switch span.SpanKind {
		case trace.SpanKindServer:
			edgeServerSpan = span
		case trace.SpanKindClient:
			edgeClientSpan = span
		}
	}
	if edgeTraceSpanCount != 2 {
		t.Errorf("Edge spans for trace %s = %d, want 2", wantTraceID, edgeTraceSpanCount)
	}
	if edgeServerSpan == nil || edgeClientSpan == nil {
		t.Fatalf("Edge spans = %#v, want inbound server and outbound client spans for %s", edgeSpans, wantTraceID)
	}
	if edgeServerSpan.Parent.SpanID() != wantRootParentID {
		t.Errorf("Edge server parent = %s, want incoming parent %s", edgeServerSpan.Parent.SpanID(), wantRootParentID)
	}
	if edgeServerSpan.SpanContext.TraceID() != edgeClientSpan.SpanContext.TraceID() {
		t.Errorf("Edge server and client trace IDs differ: %s and %s", edgeServerSpan.SpanContext.TraceID(), edgeClientSpan.SpanContext.TraceID())
	}
	if edgeClientSpan.Parent.SpanID() != edgeServerSpan.SpanContext.SpanID() {
		t.Errorf("Edge client parent = %s, want Edge server span %s", edgeClientSpan.Parent.SpanID(), edgeServerSpan.SpanContext.SpanID())
	}

	var ordersSpan *tracetest.SpanStub
	ordersTraceSpanCount := 0
	for i := range ordersSpans {
		span := &ordersSpans[i]
		if span.SpanContext.TraceID() == wantTraceID {
			ordersTraceSpanCount++
			ordersSpan = span
		}
	}
	if ordersTraceSpanCount != 1 {
		t.Errorf("Orders spans for trace %s = %d, want 1", wantTraceID, ordersTraceSpanCount)
	}
	if ordersSpan == nil {
		t.Fatalf("Orders spans = %#v, want a span for trace %s", ordersSpans, wantTraceID)
	}
	if ordersSpan.SpanKind != trace.SpanKindServer {
		t.Errorf("Orders span kind = %s, want server", ordersSpan.SpanKind)
	}
	if ordersSpan.SpanContext.TraceID() != edgeClientSpan.SpanContext.TraceID() {
		t.Errorf("Orders trace ID = %s, want Edge trace ID %s", ordersSpan.SpanContext.TraceID(), edgeClientSpan.SpanContext.TraceID())
	}
	if ordersSpan.Parent.SpanID() != edgeClientSpan.SpanContext.SpanID() {
		t.Errorf("Orders parent = %s, want Edge client span %s", ordersSpan.Parent.SpanID(), edgeClientSpan.SpanContext.SpanID())
	}
	if serviceName := resourceServiceName(ordersSpan.Resource); serviceName != "orders" {
		t.Errorf("Orders service.name = %q, want %q", serviceName, "orders")
	}
	if serviceName := resourceServiceName(edgeServerSpan.Resource); serviceName != "edge" {
		t.Errorf("Edge service.name = %q, want %q", serviceName, "edge")
	}
}

func newTestTracerProvider(
	t *testing.T,
	exporter *tracetest.InMemoryExporter,
	serviceName string,
) *sdktrace.TracerProvider {
	t.Helper()
	serviceResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)
	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(serviceResource),
	)
}

func resourceServiceName(serviceResource *resource.Resource) string {
	for _, kv := range serviceResource.Attributes() {
		if kv.Key == attribute.Key("service.name") {
			return kv.Value.AsString()
		}
	}
	return ""
}
