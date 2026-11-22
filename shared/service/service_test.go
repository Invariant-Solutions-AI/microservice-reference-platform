// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		value string
		want  slog.Level
	}{
		{value: "debug", want: slog.LevelDebug},
		{value: "info", want: slog.LevelInfo},
		{value: "warn", want: slog.LevelWarn},
		{value: "error", want: slog.LevelError},
		{value: "unknown", want: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := ParseLogLevel(tt.value); got != tt.want {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestHealthHandlers(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{name: "health", handler: Health},
		{name: "ready", handler: Ready},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
			}
		})
	}
}

func TestHTTPServerExtractsTraceContext(t *testing.T) {
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
	otel.SetTextMapPropagator(propagation.TraceContext{})

	for _, serviceName := range []string{"edge", "orders"} {
		t.Run(serviceName, func(t *testing.T) {
			exporter := tracetest.NewInMemoryExporter()
			provider := sdktrace.NewTracerProvider(
				sdktrace.WithSyncer(exporter),
				sdktrace.WithResource(resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceName(serviceName),
				)),
			)
			otel.SetTracerProvider(provider)
			t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })

			server := newHTTPServer(serviceName, ":0", http.HandlerFunc(Health))
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			req.Header.Set("traceparent", "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
			rec := httptest.NewRecorder()
			server.Handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			spans := exporter.GetSpans()
			if len(spans) != 1 {
				t.Fatalf("exported spans = %d, want 1", len(spans))
			}
			span := spans[0]
			wantTraceID, _ := trace.TraceIDFromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			wantParentID, _ := trace.SpanIDFromHex("bbbbbbbbbbbbbbbb")
			if span.SpanKind != trace.SpanKindServer {
				t.Errorf("span kind = %v, want server", span.SpanKind)
			}
			if span.SpanContext.TraceID() != wantTraceID || span.Parent.SpanID() != wantParentID {
				t.Errorf("span trace ID and parent = %s/%s, want %s/%s",
					span.SpanContext.TraceID(), span.Parent.SpanID(), wantTraceID, wantParentID)
			}
			value, ok := span.Resource.Set().Value(semconv.ServiceNameKey)
			if !ok || value.AsString() != serviceName {
				t.Errorf("service.name = %v, want %q", value, serviceName)
			}
		})
	}
}
