package oti

import (
	"context"
	"github.com/tombenke/go-12f-common/v2/must"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TODO:
// initOtlpTracerProvider Initializes an OTLP exporter, and configures the corresponding tracer provider.
func initOtlpTracerProvider(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	// Set up a trace exporter
	tracerExporter := must.MustVal(otlptracegrpc.New(ctx))
	bsp := sdktrace.NewBatchSpanProcessor(tracerExporter)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(tracerExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set the newly created tracer provider to be global
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider, nil
}

// initConsoleTracerProvider Initializes an stdout exporter, and configures the corresponding tracer provider.
func initConsoleTracerProvider(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	tracerExporter := must.MustVal(stdouttrace.New(stdouttrace.WithPrettyPrint()))

	bsp := sdktrace.NewBatchSpanProcessor(tracerExporter)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(tracerExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set the newly created tracer provider to be global
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider, nil
}
