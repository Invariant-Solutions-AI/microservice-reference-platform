// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package routing

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestReverseProxy(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotQuery string
	var gotHeader string
	var gotBody string

	upstream := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				gotQuery = r.URL.RawQuery
				gotHeader = r.Header.Get("X-Test")

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("ReadAll() error = %v", err)
				}

				gotBody = string(body)

				w.Header().Set(
					"X-Upstream",
					"orders",
				)
				w.WriteHeader(http.StatusCreated)

				_, _ = w.Write([]byte(`{"ok":true}`))
			},
		),
	)
	defer upstream.Close()

	target, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf(
			"url.Parse() error = %v",
			err,
		)
	}

	proxy := NewReverseProxy(target)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders?source=edge",
		http.NoBody,
	)
	req.Header.Set(
		"X-Test",
		"value",
	)

	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if gotMethod != http.MethodPost {
		t.Errorf(
			"method = %q, want %q",
			gotMethod,
			http.MethodPost,
		)
	}

	if gotPath != "/orders" {
		t.Errorf(
			"path = %q, want %q",
			gotPath,
			"/orders",
		)
	}

	if gotQuery != "source=edge" {
		t.Errorf(
			"query = %q, want %q",
			gotQuery,
			"source=edge",
		)
	}

	if gotHeader != "value" {
		t.Errorf(
			"X-Test = %q, want %q",
			gotHeader,
			"value",
		)
	}

	if gotBody != "" {
		t.Errorf(
			"body = %q, want empty",
			gotBody,
		)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf(
			"status = %d, want %d",
			rec.Code,
			http.StatusCreated,
		)
	}

	if got := rec.Header().Get("X-Upstream"); got != "orders" {
		t.Errorf(
			"X-Upstream = %q, want %q",
			got,
			"orders",
		)
	}

	if got := rec.Body.String(); got != `{"ok":true}` {
		t.Errorf(
			"body = %q, want %q",
			got,
			`{"ok":true}`,
		)
	}
}

func TestReverseProxySetsForwardedHeaders(t *testing.T) {
	var forwardedHost string
	var forwardedProto string

	upstream := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				forwardedHost = r.Header.Get(
					"X-Forwarded-Host",
				)
				forwardedProto = r.Header.Get(
					"X-Forwarded-Proto",
				)

				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer upstream.Close()

	target, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	proxy := NewReverseProxy(target)

	req := httptest.NewRequest(
		http.MethodGet,
		"http://api.example.com/orders",
		nil,
	)

	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if forwardedHost != "api.example.com" {
		t.Errorf(
			"X-Forwarded-Host = %q, want %q",
			forwardedHost,
			"api.example.com",
		)
	}

	if forwardedProto != "http" {
		t.Errorf(
			"X-Forwarded-Proto = %q, want %q",
			forwardedProto,
			"http",
		)
	}
}

func TestReverseProxyReturnsBadGatewayWhenUpstreamFails(
	t *testing.T,
) {
	target, err := url.Parse(
		"http://127.0.0.1:1",
	)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	proxy := NewReverseProxy(target)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadGateway,
		)
	}
}

type contextLogHandler struct {
	contexts map[string]trace.SpanContext
}

func (h *contextLogHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *contextLogHandler) Handle(ctx context.Context, record slog.Record) error {
	h.contexts[record.Message] = trace.SpanContextFromContext(ctx)
	return nil
}
func (h *contextLogHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *contextLogHandler) WithGroup(string) slog.Handler      { return h }

func TestProxyRequestLogsUseRequestContext(t *testing.T) {
	target, err := url.Parse("http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	handler := Orders(NewReverseProxy(target))
	logs := &contextLogHandler{contexts: make(map[string]trace.SpanContext)}
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(logs))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	traceID, _ := trace.TraceIDFromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	spanID, _ := trace.SpanIDFromHex("bbbbbbbbbbbbbbbb")
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req = req.WithContext(trace.ContextWithSpanContext(req.Context(), spanContext))
	handler.ServeHTTP(httptest.NewRecorder(), req)

	for _, message := range []string{
		"Original request path",
		"Trimmed request path",
		"Request path set for proxy",
		"Rewriting request for proxy",
		"Request URL set for proxy",
		"Proxy error",
	} {
		got, ok := logs.contexts[message]
		if !ok || got.TraceID() != traceID || got.SpanID() != spanID {
			t.Errorf("%q context = %v, want %s/%s", message, got, traceID, spanID)
		}
	}
}
