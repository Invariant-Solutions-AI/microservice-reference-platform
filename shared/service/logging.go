// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/otel/trace"
)

type traceHandler struct {
	slog.Handler
}

func (h traceHandler) Handle(ctx context.Context, record slog.Record) error {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, record)
}

func (h traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return traceHandler{Handler: h.Handler.WithGroup(name)}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := 0
		setOK := func() {
			if status == 0 {
				status = http.StatusOK
			}
		}
		wrapped := httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					if status == 0 && (code >= 200 || code == http.StatusSwitchingProtocols) {
						status = code
					}
					next(code)
				}
			},
			Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(b []byte) (int, error) {
					setOK()
					return next(b)
				}
			},
			Flush: func(next httpsnoop.FlushFunc) httpsnoop.FlushFunc {
				return func() {
					setOK()
					next()
				}
			},
			FlushError: func(next httpsnoop.FlushErrorFunc) httpsnoop.FlushErrorFunc {
				return func() error {
					setOK()
					return next()
				}
			},
			ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				return func(src io.Reader) (int64, error) {
					setOK()
					return next(src)
				}
			},
			WriteString: func(next httpsnoop.WriteStringFunc) httpsnoop.WriteStringFunc {
				return func(s string) (int, error) {
					setOK()
					return next(s)
				}
			},
		})
		next.ServeHTTP(wrapped, r)
		setOK()
		slog.InfoContext(r.Context(), "HTTP request completed",
			"method", r.Method,
			"status", status,
			"duration_ms", float64(time.Since(start))/float64(time.Millisecond),
		)
	})
}
