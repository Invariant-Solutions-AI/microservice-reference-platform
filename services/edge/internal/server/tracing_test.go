// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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

func TestOrdersRequestPropagatesTraceContext(t *testing.T) {
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()

	edgeExporter := tracetest.NewInMemoryExporter()
	edgeProvider := newTestTracerProvider(t, edgeExporter, "edge")
	otel.SetTracerProvider(edgeProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	ordersExporter := tracetest.NewInMemoryExporter()
	ordersProvider := newTestTracerProvider(t, ordersExporter, "orders")
	t.Cleanup(func() {
		_ = edgeProvider.Shutdown(t.Context())
		_ = ordersProvider.Shutdown(t.Context())
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
	var ordersSpanContext trace.SpanContext
	var ordersTraceparent string
	ordersHandler := otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ordersSpanContext = trace.SpanFromContext(r.Context()).SpanContext()
			ordersTraceparent = r.Header.Get("traceparent")
			w.WriteHeader(http.StatusCreated)
		}),
		"orders",
		otelhttp.WithTracerProvider(ordersProvider),
	)
	ordersServer := httptest.NewServer(ordersHandler)
	t.Cleanup(ordersServer.Close)

	target, err := url.Parse(ordersServer.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	handler := otelhttp.NewHandler(NewHandler(target), "edge")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", http.NoBody)
	req.Header.Set("traceparent", "00-11111111111111111111111111111111-2222222222222222-01")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if !ordersSpanContext.IsValid() {
		t.Fatal("Orders request did not create a valid server span")
	}

	edgeSpans := edgeExporter.GetSpans()
	var edgeServerSpan, edgeClientSpan *tracetest.SpanStub
	for i := range edgeSpans {
		span := &edgeSpans[i]
		if span.SpanKind == trace.SpanKindServer {
			edgeServerSpan = span
		}
		if span.SpanKind == trace.SpanKindClient {
			edgeClientSpan = span
		}
	}
	if edgeServerSpan == nil || edgeClientSpan == nil {
		t.Fatalf("edge spans = %#v, want inbound server and outbound client spans", edgeSpans)
	}

	wantTraceID, _ := trace.TraceIDFromHex("11111111111111111111111111111111")
	wantParentID, _ := trace.SpanIDFromHex("2222222222222222")
	if edgeServerSpan.SpanContext.TraceID() != wantTraceID {
		t.Errorf("edge trace ID = %s, want %s", edgeServerSpan.SpanContext.TraceID(), wantTraceID)
	}
	if edgeServerSpan.Parent.SpanID() != wantParentID {
		t.Errorf("edge parent span ID = %s, want %s", edgeServerSpan.Parent.SpanID(), wantParentID)
	}
	if edgeClientSpan.SpanContext.TraceID() != wantTraceID {
		t.Errorf("edge client trace ID = %s, want %s", edgeClientSpan.SpanContext.TraceID(), wantTraceID)
	}
	if edgeClientSpan.Parent.SpanID() != edgeServerSpan.SpanContext.SpanID() {
		t.Errorf("edge client parent = %s, want edge server span %s", edgeClientSpan.Parent.SpanID(), edgeServerSpan.SpanContext.SpanID())
	}
	if ordersSpanContext.TraceID() != wantTraceID {
		t.Errorf("Orders trace ID = %s, want %s", ordersSpanContext.TraceID(), wantTraceID)
	}
	if ordersSpanContext.TraceID() != edgeClientSpan.SpanContext.TraceID() {
		t.Errorf("Orders trace ID = %s, want edge client trace ID %s", ordersSpanContext.TraceID(), edgeClientSpan.SpanContext.TraceID())
	}
	injectedHeader := make(http.Header)
	injectedHeader.Set("traceparent", ordersTraceparent)
	injectedContext := propagation.TraceContext{}.Extract(
		context.Background(),
		propagation.HeaderCarrier(injectedHeader),
	)
	injectedSpanContext := trace.SpanContextFromContext(injectedContext)
	if !injectedSpanContext.IsValid() {
		t.Fatal("Orders request did not receive a valid traceparent header")
	}
	if injectedSpanContext.TraceID() != wantTraceID {
		t.Errorf("injected trace ID = %s, want %s", injectedSpanContext.TraceID(), wantTraceID)
	}
	if injectedSpanContext.SpanID() != edgeClientSpan.SpanContext.SpanID() {
		t.Errorf("injected parent span ID = %s, want edge client span %s", injectedSpanContext.SpanID(), edgeClientSpan.SpanContext.SpanID())
	}

	ordersSpans := ordersExporter.GetSpans()
	if len(ordersSpans) != 1 {
		t.Fatalf("Orders spans = %d, want 1", len(ordersSpans))
	}
	if ordersSpans[0].Parent.SpanID() != edgeClientSpan.SpanContext.SpanID() {
		t.Errorf("Orders parent span ID = %s, want edge client span %s", ordersSpans[0].Parent.SpanID(), edgeClientSpan.SpanContext.SpanID())
	}
	if serviceName := resourceServiceName(ordersSpans[0].Resource); serviceName != "orders" {
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
