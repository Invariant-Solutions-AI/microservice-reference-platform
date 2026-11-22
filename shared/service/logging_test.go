// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceHandler(t *testing.T) {
	traceID, _ := trace.TraceIDFromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	spanID, _ := trace.SpanIDFromHex("bbbbbbbbbbbbbbbb")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))

	for _, tt := range []struct {
		name string
		ctx  context.Context
		want bool
	}{
		{name: "valid unsampled span", ctx: ctx, want: true},
		{name: "no span", ctx: context.Background(), want: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(traceHandler{Handler: slog.NewJSONHandler(&output, nil)}).
				With("service", "edge")
			logger.InfoContext(tt.ctx, "request")

			var record map[string]any
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if record["service"] != "edge" || record["msg"] != "request" {
				t.Errorf("log record = %v", record)
			}
			gotTrace, hasTrace := record["trace_id"]
			gotSpan, hasSpan := record["span_id"]
			if hasTrace != tt.want || hasSpan != tt.want {
				t.Fatalf("trace fields in log = %v, want present: %t", record, tt.want)
			}
			if tt.want && (gotTrace != traceID.String() || gotSpan != spanID.String()) {
				t.Errorf("trace fields = %v/%v, want %s/%s", gotTrace, gotSpan, traceID, spanID)
			}
		})
	}
}

func TestNewLoggerDeploymentMetadata(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
		want   map[string]string
	}{
		{
			name: "configured",
			values: map[string]string{
				"SERVICE_VERSION":        "1.2.3",
				"DEPLOYMENT_ENVIRONMENT": "production",
				"POD_UID":                "pod-uid",
				"POD_NAME":               "orders-123",
				"POD_NAMESPACE":          "shop",
				"NODE_NAME":              "node-1",
			},
			want: map[string]string{
				"service_version":        "1.2.3",
				"deployment_environment": "production",
				"service_instance_id":    "pod-uid",
				"k8s_pod_name":           "orders-123",
				"k8s_namespace_name":     "shop",
				"k8s_node_name":          "node-1",
			},
		},
		{name: "not configured"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, name := range []string{
				"SERVICE_VERSION", "DEPLOYMENT_ENVIRONMENT", "POD_UID",
				"POD_NAME", "POD_NAMESPACE", "NODE_NAME",
			} {
				t.Setenv(name, tt.values[name])
			}

			previousLogger := slog.Default()
			t.Cleanup(func() { slog.SetDefault(previousLogger) })
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = reader.Close()
				_ = writer.Close()
			})
			previousStdout := os.Stdout
			os.Stdout = writer
			NewLogger("orders")
			os.Stdout = previousStdout
			slog.Info("started")
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			output, err := io.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}
			var record map[string]any
			if err := json.Unmarshal(output, &record); err != nil {
				t.Fatal(err)
			}
			if record["service"] != "orders" {
				t.Errorf("service = %v, want orders", record["service"])
			}
			for _, key := range []string{
				"service_version", "deployment_environment", "service_instance_id",
				"k8s_pod_name", "k8s_namespace_name", "k8s_node_name",
			} {
				got, exists := record[key]
				want, configured := tt.want[key]
				if exists != configured || (configured && got != want) {
					t.Errorf("%s = %v (present %t), want %q (present %t)",
						key, got, exists, want, configured)
				}
			}
		})
	}
}

func TestRequestLogUsesServerSpan(t *testing.T) {
	previousLogger := slog.Default()
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() {
		_ = provider.Shutdown(t.Context())
		slog.SetDefault(previousLogger)
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = reader.Close()
		_ = writer.Close()
	})
	previousStdout := os.Stdout
	os.Stdout = writer
	level := NewLogger("orders")
	os.Stdout = previousStdout
	level.Set(slog.LevelDebug)

	server := newHTTPServer("orders", ":0", http.HandlerFunc(Health))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("traceparent", "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
	server.Handler.ServeHTTP(httptest.NewRecorder(), req)

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(output, &record); err != nil {
		t.Fatal(err)
	}
	if record["service"] != "orders" {
		t.Errorf("service = %v, want orders", record["service"])
	}
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported spans = %d, want 1", len(spans))
	}
	if record["trace_id"] != spans[0].SpanContext.TraceID().String() ||
		record["span_id"] != spans[0].SpanContext.SpanID().String() {
		t.Errorf("log trace fields = %v/%v, want server span %s/%s",
			record["trace_id"], record["span_id"],
			spans[0].SpanContext.TraceID(), spans[0].SpanContext.SpanID())
	}
}

func TestTraceHandlerWithGroupAndLevel(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(traceHandler{Handler: slog.NewJSONHandler(&output, nil)}).WithGroup("request").With("method", "GET")
	logger.DebugContext(context.Background(), "filtered")
	if output.Len() != 0 {
		t.Fatal("debug log should be filtered")
	}

	traceID, _ := trace.TraceIDFromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	spanID, _ := trace.SpanIDFromHex("bbbbbbbbbbbbbbbb")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))
	logger.InfoContext(ctx, "request")
	var record struct {
		Request struct {
			Method  string `json:"method"`
			TraceID string `json:"trace_id"`
			SpanID  string `json:"span_id"`
		} `json:"request"`
	}
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record.Request.Method != "GET" || record.Request.TraceID != traceID.String() || record.Request.SpanID != spanID.String() {
		t.Errorf("grouped log record = %+v", record)
	}
}
