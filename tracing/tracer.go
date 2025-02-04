package tracing

import (
	"context"
	"os"
	"regexp"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

	logger "github.com/tombenke/go-12f-common/log"

	"github.com/tombenke/go-12f-common/middleware/model"
)

const (
	TracerVersion = "0.1.0"
)

type ErrorHandler struct {
	log logger.FieldLogger
}

func (e *ErrorHandler) Handle(err error) {
	e.log.WithField(logger.KeyError, err).Error("OTEL ERROR")
}

var errorHandler = &ErrorHandler{} //nolint:gochecknoglobals // local once
var onceSetOtel sync.Once          //nolint:gochecknoglobals // local once
var onceBodySetOtel = func() {     //nolint:gochecknoglobals // local once
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	))
	otel.SetErrorHandler(errorHandler)
	// TODO logr --> slog, see: https://github.com/go-logr/logr/pull/196
	//otel.SetLogger(*errorHandler.log)
}

func SetErrorHandlerLogger(log logger.FieldLogger) {
	errorHandler.log = log
}

const (
	StateKeyClientCommand = "client_command"
	//StateKeyOidcCommand   = "oidc_command"
	SpanKeyComponent = "component"
	SpanKeyService   = "service"
	SpanKeyInstance  = "instance"
)

var (
	SpanKeyComponentValue string
)

func InitTracer(exporter sdktrace.SpanExporter, sampler sdktrace.Sampler, buildinfo model.BuildInfo, service string, instance string, command string, log logger.FieldLogger) *sdktrace.TracerProvider {
	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	// semconv keys are defined in https://github.com/open-telemetry/opentelemetry-specification/tree/main/semantic_conventions/trace
	attrs := []attribute.KeyValue{
		semconv.ServiceNamespaceKey.String(buildinfo.AppName()),
		semconv.ServiceNameKey.String(service),
		semconv.ServiceInstanceIDKey.String(instance),
		semconv.ServiceVersionKey.String(buildinfo.Version()),
		attribute.Int("attrID", os.Getpid()),
	}
	SpanKeyComponentValue = buildinfo.AppName()
	if command != "" {
		attrs = append(attrs, attribute.String(StateKeyClientCommand, command))
	}
	providerOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attrs...)),
	}
	if exporter != nil {
		providerOptions = append(providerOptions, sdktrace.WithBatcher(exporter))
	}
	tp := sdktrace.NewTracerProvider(providerOptions...)

	if errorHandler.log == nil {
		errorHandler.log = log
	}
	onceSetOtel.Do(onceBodySetOtel)

	return tp
}

func OtlpProvider(ctx context.Context, options ...otlptracehttp.Option) (sdktrace.SpanExporter, error) {
	if len(options) == 0 {
		return nil, nil
	}
	return otlptracehttp.New(ctx, options...)
}

var invalidTracestateValueRe = regexp.MustCompile(`[^\x20-\x2b\x2d-\x3c\x3e-\x7e]`)

func EncodeTracestateValue(value string) string {
	return invalidTracestateValueRe.ReplaceAllString(strings.TrimSpace(value), "_")
}

func Version() string {
	return TracerVersion
}

// SemVersion is the semantic version to be supplied to tracer/meter creation.
func SemVersion() string {
	return "semver:" + Version()
}
