// Package oti provides the basic features required for Open-Telemetry instrumentation
package oti

import (
	"context"
	"os"
	"time"

	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	export "go.opentelemetry.io/otel/sdk/export/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

const (
	ActorLatency = attribute.Key("actor.latency")
)

var (
	logger = log.Logger
)

func MergeAttributes(arrays ...[]attribute.KeyValue) []attribute.KeyValue {
	length := 0
	for _, a := range arrays {
		length += len(a)
	}

	attributes := make([]attribute.KeyValue, 0, length)

	for _, a := range arrays {
		attributes = append(attributes, a...)
	}

	return attributes
}

func RecordMetrics(ctx context.Context, attributes []attribute.KeyValue, measurements ...metric.Measurement) {
	if len(measurements) > 0 {
		global.Meter("").RecordBatch(ctx, attributes, measurements...)
	}
}

// OTI holds the required objects to make OTEL instrumentation working in a specific library of application
type OTI struct {
	tracerProvider *sdktrace.TracerProvider
	pusher         *controller.Controller
	config         Config

	// Ctx is a central context used for initializing tracer and meter providers
	Ctx context.Context

	// Tracer is a global tracer
	Tracer trace.Tracer

	// Meter is a global meter
	Meter metric.Meter
}

// NewOTI creates a new OTEL instrumentation configuration incl. the global providers
func NewOTI(config Config) *OTI {
	var oti OTI
	// Create an instance of a Go context.
	// It will be used later to store some important data.
	oti.Ctx = context.Background()
	oti.config = config

	switch config.Exporter.Type {
	case OtelGrpcExporter:
		// 1. Create exporter(s)
		exporter := oti.newOtlpGrpcExporter(config.Exporter.CollectorURL)

		// When using OpenTelemetry, it’s a good practice to set a global tracer provider
		// and a global meter provider.
		// Doing so will make it easier for libraries and other dependencies
		// that use the OpenTelemetry API to easily discover the SDK, and emit telemetry data.
		// In addition, you’ll want to configure context propagation options.
		// Context propagation allows for OpenTelemetry to share values across multiple services
		// - this includes trace identifiers, which ensure that all spans for a single request
		// are part of the same trace, as well as baggage, which are arbitrary key/value pairs
		// that you can use to pass observability data between services
		// (for example, sharing a customer ID from one service to the next).

		// 2. Create a global Tracer Provider
		oti.setupGlobalTracerProvider(oti.Ctx, exporter)

		// 3. Create a global Meter Provider
		oti.setupGlobalMeterProvider(oti.Ctx, exporter)

	case StdoutExporter:
		exporter := oti.newStdoutExporter()
		oti.setupGlobalTracerProvider(oti.Ctx, exporter)
		oti.setupGlobalMeterProvider(oti.Ctx, exporter)
	}

	oti.Tracer = otel.Tracer("")
	oti.Meter = global.Meter("")

	// 4. Create global Propagator
	oti.setupGlobalCompositeTextPropagator()

	return &oti
}

// Shutdown shuts down the OTEL instrumentation configuration
func (oti *OTI) Shutdown() {
	// Handle this error in a sensible manner where possible
	if err := oti.tracerProvider.Shutdown(oti.Ctx); err != nil {
		logger.Errorf("%s", err.Error())
	}

	// Handle this error in a sensible manner where possible
	if err := oti.pusher.Stop(oti.Ctx); err != nil {
		logger.Errorf("%s", err.Error())
	}

}

// newStdoutExporter Creates an stdout exporter.
//
// The SDK requires an exporter to be created.
// Exporters are packages that allow telemetry data to be emitted somewhere
// - either to the console (which is what we’re doing here),
// or to a remote system or collector for further analysis and/or enrichment.
// OpenTelemetry supports a variety of exporters through its ecosystem
// including popular open source tools like Jaeger, Zipkin, and Prometheus.
//
// This function creates a new console exporter with basic options
// WithPrettyPrint formats the text nicely when its printed, so that it’s easier for humans to read.
func (oti *OTI) newStdoutExporter() *stdout.Exporter {
	exporter, err := stdout.NewExporter(
		stdout.WithPrettyPrint(),
	)
	if err != nil {
		logger.Fatalf("failed to initialize stdout export pipeline: %v", err)
	}

	return exporter
}

// newOtlpGrpcExporter Creates an OTLP GRPC exporter
func (oti *OTI) newOtlpGrpcExporter(otelAgentAddr string) *otlp.Exporter {
	exporter, err := otlp.NewExporter(oti.Ctx, otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(otelAgentAddr),
	))
	if err != nil {
		logger.Fatalf("failed to initialize OTLP-GRPC export pipeline: %v", err)
	}

	return exporter
}

// setupGlobalTracerProvider Creates a Trace Provider, then sets it to be global
// A trace is a type of telemetry that represents work being done by a service.
// In a distributed system, a trace can be thought of as a ‘stack trace’,
// showing the work being done by each service as well as the upstream and downstream calls
// that its making to other services.
// OpenTelemetry requires a trace provider to be initialized in order to generate traces.
// A trace provider can have multiple span processors, which are components
// that allow for span data to be modified or exported after it’s created.
//
// Create a new batch span processor,
// a type of span processor that batches up multiple spans over a period of time,
// that writes to the exporter we created in the previous step.
func (oti *OTI) setupGlobalTracerProvider(ctx context.Context, exporter sdktrace.SpanExporter) {
	// TODO: Move to its final place
	res, _ := resource.New(ctx,
		resource.WithAttributes(oti.Attributes()...),
	)

	var spanProcessor sdktrace.SpanProcessor
	switch oti.config.SpanProcessorType {
	case BatchSpanProcessor:
		spanProcessor = sdktrace.NewBatchSpanProcessor(exporter)
	case SimpleSpanProcessor:
		fallthrough
	default:
		spanProcessor = sdktrace.NewSimpleSpanProcessor(exporter)
	}

	var sampler sdktrace.Sampler
	switch oti.config.Sampling.Type {
	case NeverSampling:
		sampler = sdktrace.NeverSample()
	case AlwaysSampling:
		sampler = sdktrace.AlwaysSample()
	case ParentBasedSampling:
		panic("ParentBasedSampling type is not implemented!")
		// TODO
		//sampler = sdktrace.ParentBasedSample(...)
	case RatioBasedSampling:
		sampler = sdktrace.TraceIDRatioBased(oti.config.Sampling.Ratio)
	}

	oti.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(spanProcessor),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(oti.tracerProvider)
}

// setupGlobalMeterProvider
//
// OpenTelemetry requires a meter provider to be initialized
// in order to create instruments that will generate metrics.
// The way metrics are exported depends on the used system.
// For example, prometheus uses a pull model, while OTLP uses a push model.
// In this document we use an stdout exporter which uses the latter.
// Thus we need to create a push controller that will periodically push
// the collected metrics to the exporter.
//
// Creates a controller that uses a basic processor to aggregate and process metrics
// that are then sent to the exporter.
// The basic processor here uses a simple aggregator selector
// that decides what kind of an aggregator to use
// to aggregate measurements from a specific instrument.
// The processor also uses the exporter to learn how to prepare the aggregated measurements
// for the exporter to consume.
// The controller will periodically push aggregated measurements to the exporter.
func (oti *OTI) setupGlobalMeterProvider(ctx context.Context, exporter export.Exporter) {
	oti.pusher = controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exporter,
		),
		controller.WithExporter(exporter),
		controller.WithCollectPeriod(5*time.Second),
	)

	err := oti.pusher.Start(ctx)
	if err != nil {
		logger.Fatalf("failed to initialize metric controller: %v", err)
	}

	global.SetMeterProvider(oti.pusher.MeterProvider())
}

// setupGlobalCompositeTextMapPropagator Creates a new CompositeTextMapPropagator and makes it global
func (oti *OTI) setupGlobalCompositeTextPropagator() {
	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)
}

func (oti *OTI) Attributes() []attribute.KeyValue {
	hostname, err := os.Hostname()
	must.Must(err)

	attributes := make([]attribute.KeyValue, 2, 8)
	attributes[0] = semconv.ServiceNameKey.String(string(oti.config.ServiceName))
	attributes[1] = semconv.ServiceInstanceIDKey.String(hostname)
	if oti.config.ServiceNamespace != "" {
		attributes = append(attributes, semconv.ServiceNamespaceKey.String(string(oti.config.ServiceNamespace)))
	}

	return attributes
}
