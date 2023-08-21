package oti

import (
	"fmt"
	"strings"
)

const (
	// DefaultServiceName is the default name of the service instrumented
	DefaultServiceName = "Undefined"

	// DefaultCollectorURL is the default URL of the OTEL collector that the Otel-GRPC exporter uses
	DefaultCollectorURL = "localhost:4317"

	// StdoutExporter is an ExporterType enum value for the OTEL stdout exporter
	StdoutExporter = ExporterType(0)
	// OtelGrpcExporter is an ExporterType enum value for the OTEL GRPC exporter
	OtelGrpcExporter = ExporterType(1)
	// DefaultExporterType is the default exporter type enum value
	DefaultExporterType = StdoutExporter

	// NeverSampling is a TraceSamplingType enum value that switches sampling off
	NeverSampling = TraceSamplingType(0)
	// ParentBasedSampling is a TraceSamplingType enum value that switches to parent-based sampling mode
	ParentBasedSampling = TraceSamplingType(1)
	// RatioBasedSampling is a TraceSamplingType enum value that switches to ratio-based sampling mode
	RatioBasedSampling = TraceSamplingType(2)
	// AlwaysSampling is a TraceSamplingType enum value that switches to always do sampling
	AlwaysSampling = TraceSamplingType(3)
	// DefaultTraceSamplingType defines the default TraceSamplingType
	DefaultTraceSamplingType = NeverSampling
	// DefaultTraceSamplingRatio defines the default sampling ratio, if the sampling type id ratio-based
	DefaultTraceSamplingRatio = 1.

	// SimpleSpanProcessor is a SpanProcessorType enum value that makes span processor to work in single mode
	SimpleSpanProcessor = SpanProcessorType(0)
	// BatchSpanProcessor is a SpanProcessorType enum value that makes span processor to work in batch mode
	BatchSpanProcessor = SpanProcessorType(1)
	// DefaultSpanProcessorType defines the default span processor type
	DefaultSpanProcessorType = SimpleSpanProcessor
)

// ServiceName is the type of the name of the tracing service
type ServiceName string

// ServiceNamespace is the name of the namespace that the service belongs to.
type ServiceNamespace string

// ExporterType is the type of the exporter used for sending telemetry data
type ExporterType int

// TraceSamplingType is the type to select from the trace sampling modes
type TraceSamplingType int

// SpanProcessorType is the type to select from the span processor modes
type SpanProcessorType int

// Config holds the configuration parameters for the Open Telemetry instrumentation
type Config struct {
	// ServiceName is the name of the service that collects metrics and tracing
	ServiceName ServiceName
	// ServiceNamespace is the name of the namespace that the service belongs to.
	ServiceNamespace ServiceNamespace
	// Exporter holds the configuration parameters for the actual exporter of the OTEL instrumentation
	Exporter ExporterConfig
	// Sampling holds the configuration parameters for the actual sampling mode of the OTEL instrumentation
	Sampling SamplingConfig
	// SpanProcessorType defines which span processor type is used by the actual configuration
	SpanProcessorType SpanProcessorType
}

// ExporterConfig holds the configuration parameters for the actual exporter of the OTEL instrumentation
type ExporterConfig struct {
	// Type defines the type of exporter used
	Type ExporterType
	// CollectorURL defines the telemetry data collector sidecar or agent in case the exporter-type is OTEL-GRPC
	CollectorURL string
}

// SamplingConfig holds the configuration parameters for the actual sampling mode of the OTEL instrumentation
type SamplingConfig struct {
	// Type defines the type of trace sampling mode
	Type TraceSamplingType
	// Ratio defines the tracing ratio, if the tracing type is RatioBasedSampling
	Ratio float64
}

// NewConfig creates a new configuration setup for OTEL instrumentation with the dafalut values,
// that can be overwritten by the options
func NewConfig(options ...Option) *Config {
	cfg := Config{
		ServiceName: DefaultServiceName,
		Exporter: ExporterConfig{
			Type:         DefaultExporterType,
			CollectorURL: DefaultCollectorURL,
		},
		Sampling: SamplingConfig{
			Type:  DefaultTraceSamplingType,
			Ratio: -1,
		},
		SpanProcessorType: DefaultSpanProcessorType,
	}

	for _, option := range options {
		option.Apply(&cfg)
	}

	return &cfg
}

// Option is an interface for OTEL configuration options
type Option interface {
	// Apply applies the configuration parameter value to the specific property of the configuration object
	Apply(*Config)
}

// Apply sets the service name parameter of the configuration object
func (t ServiceName) Apply(cfg *Config) {
	cfg.ServiceName = t
}

// WithServiceName converts a string-type service name to a ServiceName type option
func WithServiceName(serviceName string) ServiceName {
	return ServiceName(serviceName)
}

// Apply sets the exporter parameters of the configuration object
func (t ExporterConfig) Apply(cfg *Config) {
	cfg.Exporter.Type = t.Type
	cfg.Exporter.CollectorURL = t.CollectorURL
}

// WithStdoutExporter returns with an STDOUT ExporterConfig option
func WithStdoutExporter() ExporterConfig {
	return ExporterConfig{
		Type:         StdoutExporter,
		CollectorURL: "",
	}
}

// WithOtelGrpcCollectorExporter returns with an OTEL GRPC ExporterConfig option, incl. the type and the URL
func WithOtelGrpcCollectorExporter(collectorURL string) ExporterConfig {
	return ExporterConfig{
		Type:         OtelGrpcExporter,
		CollectorURL: collectorURL,
	}
}

// Apply sets the sampling parameters of the configuration object
func (t SamplingConfig) Apply(cfg *Config) {
	cfg.Sampling.Type = t.Type
	cfg.Sampling.Ratio = t.Ratio
}

// WithNeverSampling returns with a SamplingConfig that switches sampling off
func WithNeverSampling() SamplingConfig {
	return SamplingConfig{
		Type:  NeverSampling,
		Ratio: -1,
	}
}

// WithRatioBasedSampling returns a SamplingConfig that sets the sampling to ratio-based with the given ratio value
func WithRatioBasedSampling(ratio float64) SamplingConfig {
	return SamplingConfig{
		Type:  RatioBasedSampling,
		Ratio: ratio,
	}
}

// WithParentBasedSampling returns a SamplingConfig that sets the sampling to be parent-based
func WithParentBasedSampling() SamplingConfig {
	return SamplingConfig{
		Type:  ParentBasedSampling,
		Ratio: -1,
	}
}

// WithAlwaysSampling returns with a SamplingConfig that switches the app to do always sampling
func WithAlwaysSampling() SamplingConfig {
	return SamplingConfig{
		Type:  AlwaysSampling,
		Ratio: 1,
	}
}

// Apply sets the span processor parameter of the configuration object
func (t SpanProcessorType) Apply(cfg *Config) {
	cfg.SpanProcessorType = t
}

// WithSimpleSpanProcessor returns with a SpanProcessorType that selects the simple-type span processor
func WithSimpleSpanProcessor() SpanProcessorType {
	return SimpleSpanProcessor
}

// WithBatchSpanProcessor returns with a SpanProcessorType that selects the batch-type span processor
func WithBatchSpanProcessor() SpanProcessorType {
	return BatchSpanProcessor
}

// WithExporter creates an ExporterType value selected by one of the following string values: `STDOUT`, `OTELGRPC`.
// It also returns an error, in case a wrong string value is used.
func WithExporter(exporterType string) (ExporterType, error) {
	switch strings.ToUpper(exporterType) {
	case "STDOUT":
		return StdoutExporter, nil
	case "OTELGRPC":
		return OtelGrpcExporter, nil
	default:
		return DefaultExporterType, fmt.Errorf("wrong otel exporter-type: '%s'", exporterType)
	}
}

// String return with the ExporterType value in string representation
func (t ExporterType) String() string {
	switch t {
	case StdoutExporter:
		return "STDOUT"
	case OtelGrpcExporter:
		return "OTELGRPC"
	default:
		return "UNKNOWN_EXPORTER_TYPE"
	}
}

// WithSpanProcessor creates an SpanProcessorType value selected by one of the following string values:
// `SIMPLE`, `BATCH`. It also returns an error, in case a wrong string value is used.
func WithSpanProcessor(spanProcessorType string) (SpanProcessorType, error) {
	switch strings.ToUpper(spanProcessorType) {
	case "SIMPLE":
		return SimpleSpanProcessor, nil
	case "BATCH":
		return BatchSpanProcessor, nil
	default:
		return DefaultSpanProcessorType, fmt.Errorf("wrong otel span-processor-type: '%s'", spanProcessorType)
	}
}

// String return with the SpanProcessorType value in string representation
func (t SpanProcessorType) String() string {
	switch t {
	case SimpleSpanProcessor:
		return "SIMPLE"
	case BatchSpanProcessor:
		return "BATCH"
	default:
		return "UNKNOWN_SPAN_PROCESSOR_TYPE"
	}
}

// WithTraceSampling creates an TraceSamplingType value selected by one of the following string values:
// `NEVER`, `PARENTBASED`, `RATIOBASED`, `ALWAYS`. It also returns an error, in case a wrong string value is used.
func WithTraceSampling(traceSamplingType string) (TraceSamplingType, error) {
	switch strings.ToUpper(traceSamplingType) {
	case "NEVER":
		return NeverSampling, nil
	case "PARENTBASED":
		return ParentBasedSampling, nil
	case "RATIOBASED":
		return RatioBasedSampling, nil
	case "ALWAYS":
		return AlwaysSampling, nil
	default:
		return DefaultTraceSamplingType, fmt.Errorf("wrong otel trace-sampling-type: '%s'", traceSamplingType)
	}
}

// String return with the TraceSamplingType value in string representation
func (t TraceSamplingType) String() string {
	switch t {
	case NeverSampling:
		return "NEVER"
	case ParentBasedSampling:
		return "PARENTBASED"
	case RatioBasedSampling:
		return "RATIOBASED"
	case AlwaysSampling:
		return "ALWAYS"
	default:
		return "UNKNOWN_TRACE_SAMPLING_TYPE"
	}
}
