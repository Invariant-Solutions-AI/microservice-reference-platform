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

			var output bytes.Buffer
			logger, _ := newLogger("orders", &output)
			logger.Info("started")
			var record map[string]any
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
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

	var output bytes.Buffer
	logger, _ := newLogger("orders", &output)
	slog.SetDefault(logger)

	server := newHTTPServer("orders", ":0", http.HandlerFunc(Health))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("traceparent", "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
	server.Handler.ServeHTTP(httptest.NewRecorder(), req)

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["service"] != "orders" {
		t.Errorf("service = %v, want orders", record["service"])
	}
	if record["msg"] != "HTTP request completed" || record["status"] != float64(http.StatusOK) {
		t.Errorf("unexpected completion log = %v", record)
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

func TestLoggerMinimumLevel(t *testing.T) {
	var output bytes.Buffer
	logger, level := newLogger("edge", &output)
	logger.Debug("hidden")
	logger.Info("visible")

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["msg"] != "visible" || record["level"] != "INFO" {
		t.Fatalf("default-level output = %v", record)
	}

	output.Reset()
	t.Setenv("LOG_LEVEL", "debug")
	level.Set(ParseLogLevel(os.Getenv("LOG_LEVEL")))
	logger.Debug("debug enabled")
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["msg"] != "debug enabled" || record["level"] != "DEBUG" {
		t.Errorf("debug-level output = %v", record)
	}

	output.Reset()
	level.Set(slog.LevelWarn)
	logger.Info("hidden")
	logger.Warn("warning")
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["msg"] != "warning" || record["level"] != "WARN" {
		t.Errorf("warn-level output = %v", record)
	}
}

func TestRequestCompletionStatus(t *testing.T) {
	for _, tt := range []struct {
		name    string
		handler http.HandlerFunc
		status  int
	}{
		{name: "explicit", handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}, status: http.StatusBadRequest},
		{name: "implicit write", handler: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}, status: http.StatusOK},
		{name: "empty response", handler: func(http.ResponseWriter, *http.Request) {}, status: http.StatusOK},
		{name: "flush", handler: func(w http.ResponseWriter, _ *http.Request) {
			w.(http.Flusher).Flush()
		}, status: http.StatusOK},
		{name: "first final status", handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusInternalServerError)
		}, status: http.StatusCreated},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			logger, _ := newLogger("edge", &output)
			previousLogger := slog.Default()
			slog.SetDefault(logger)
			t.Cleanup(func() { slog.SetDefault(previousLogger) })

			rec := httptest.NewRecorder()
			logRequests(tt.handler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			var record map[string]any
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if rec.Code != tt.status || record["status"] != float64(tt.status) {
				t.Errorf("response status = %d, logged status = %v, want %d", rec.Code, record["status"], tt.status)
			}
			if duration, ok := record["duration_ms"].(float64); !ok || duration < 0 {
				t.Errorf("duration_ms = %v, want nonnegative number", record["duration_ms"])
			}
			if record["method"] != http.MethodGet || record["msg"] != "HTTP request completed" {
				t.Errorf("unexpected completion log = %v", record)
			}
			if _, ok := record["trace_id"]; ok {
				t.Errorf("unexpected trace_id in untraced log: %v", record)
			}
		})
	}
}

func TestRequestCompletionAfterInformationalResponse(t *testing.T) {
	var output bytes.Buffer
	logger, _ := newLogger("edge", &output)
	previousLogger := slog.Default()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	done := make(chan struct{})
	handler := newHTTPServer("edge", ":0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusEarlyHints)
		w.WriteHeader(http.StatusCreated)
	})).Handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		close(done)
	}))
	t.Cleanup(server.Close)
	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	<-done
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated || record["status"] != float64(http.StatusCreated) {
		t.Errorf("response status = %d, logged status = %v, want %d", resp.StatusCode, record["status"], http.StatusCreated)
	}
}

type optionalResponseWriter struct {
	*httptest.ResponseRecorder
}

func (w *optionalResponseWriter) FlushError() error {
	w.ResponseRecorder.Flush()
	return nil
}

func (w *optionalResponseWriter) WriteString(s string) (int, error) {
	return w.ResponseRecorder.Write([]byte(s))
}

func TestRequestCompletionWithOptionalWriterMethods(t *testing.T) {
	for _, tt := range []struct {
		name  string
		write func(http.ResponseWriter)
	}{
		{name: "flush error", write: func(w http.ResponseWriter) {
			_ = http.NewResponseController(w).Flush()
		}},
		{name: "write string", write: func(w http.ResponseWriter) {
			_, _ = io.WriteString(w, "ok")
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			logger, _ := newLogger("edge", &output)
			previousLogger := slog.Default()
			slog.SetDefault(logger)
			t.Cleanup(func() { slog.SetDefault(previousLogger) })

			rec := &optionalResponseWriter{ResponseRecorder: httptest.NewRecorder()}
			handler := logRequests(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				tt.write(w)
				w.WriteHeader(http.StatusInternalServerError)
			}))
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			var record map[string]any
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusOK || record["status"] != float64(http.StatusOK) {
				t.Errorf("response status = %d, logged status = %v, want 200", rec.Code, record["status"])
			}
		})
	}
}
