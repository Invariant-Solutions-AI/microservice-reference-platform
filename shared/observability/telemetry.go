// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package observability

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Setup configures OpenTelemetry tracing for the current process.
//
// The returned shutdown function must be called before the process exits.
func Setup(
	ctx context.Context,
	serviceName string,
) (func(context.Context) error, error) {
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"create OTLP trace exporter: %w",
			err,
		)
	}

	serviceResource, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create OpenTelemetry resource: %w",
			err,
		)
	}

	mergedResource, err := resource.Merge(
		resource.Default(),
		serviceResource,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"merge OpenTelemetry resources: %w",
			err,
		)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(mergedResource),
	)

	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	shutdown := func(ctx context.Context) error {
		var shutdownErr error

		shutdownErr = errors.Join(
			shutdownErr,
			tracerProvider.ForceFlush(ctx),
		)
		shutdownErr = errors.Join(
			shutdownErr,
			tracerProvider.Shutdown(ctx),
		)

		return shutdownErr
	}

	return shutdown, nil
}
