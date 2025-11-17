package oti

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	logger "github.com/tombenke/go-12f-common/v2/log"
)

type OtelErrorHandler struct {
	log *slog.Logger
}

func (e *OtelErrorHandler) Handle(err error) {
	e.log.Error("OTEL ERROR", logger.FieldError, err)
}

const (
	SpanKeyComponent = attribute.Key("app")
	SpanKeyService   = semconv.ServiceNameKey
	SpanKeyInstance  = attribute.Key("instance")

	TraceparentHeader = "traceparent"
)

var (
	otelErrorHandler  atomic.Pointer[OtelErrorHandler] //nolint:gochecknoglobals // local once
	onceSetOtelGlobal = sync.OnceFunc(func() {         //nolint:gochecknoglobals // local once
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
		))
		otel.SetErrorHandler(otelErrorHandler.Load())
	})
)

// TODO:
// initTracerProvider Initializes an OTLP exporter, and configures the corresponding tracer provider.
func initTracerProvider(ctx context.Context, tracerExporter sdktrace.SpanExporter, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	// Set up a trace exporter
	bsp := sdktrace.NewBatchSpanProcessor(tracerExporter)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(tracerExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	if eh := otelErrorHandler.Load(); eh == nil {
		eh = &OtelErrorHandler{log: logger.GetFromContextOrDefault(ctx)}
		otelErrorHandler.Store(eh)
	}
	onceSetOtelGlobal()

	return tracerProvider, nil
}
