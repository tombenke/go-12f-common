package otel

import (
	"flag"

	"github.com/tombenke/go-12f-common/env"
)

const (
	serviceNameHelp    = "The value of the service.name resource attribute"
	serviceNameEnvVar  = "OTEL_SERVICE_NAME"
	ServiceNameDefault = "undefined"

	otelTracesSamplerHelp    = "Sampler to be used for traces, e.g. (always_on | always_off | traceidratio | parentbased_always_on | parentbased_always_off | parentbased_traceidratio | parentbased_jaeger_remote | jaeger_remote | xray)"
	otelTracesSamplerEnvVar  = "OTEL_TRACES_SAMPLER"
	OtelTracesSamplerDefault = "always_off"

	otelTracesSamplerArg        = "String value to be used as the sampler argument"
	otelTracesSamplerArgEnvVar  = "OTEL_TRACES_SAMPLER_ARG"
	OtelTracesSamplerArgDefault = ""

	otelTracesExporterHelp    = "Trace exporter to be used (otlp | zipkin | none)"
	otelTracesExporterEnvVar  = "OTEL_TRACES_EXPORTER"
	OtelTracesExporterDefault = "otlp"

	otelMetricsExporterHelp    = "Metrics exporter to be used (otlp | prometheus | none)"
	otelMetricsExporterEnvVar  = "OTEL_METRICS_EXPORTER"
	OtelMetricsExporterDefault = "otlp"

	otelLogsExporterHelp    = "Logs exporter to be used (otlp | none)"
	otelLogsExporterEnvVar  = "OTEL_LOGS_EXPORTER"
	OtelLogsExporterDefault = "otlp"
)

// Config holds the configuration parameters for the Open Telemetry instrumentation
type Config struct {
	// ServiceName is the name of the service that collects metrics and tracing
	ServiceName string

	// Sampler to be used for traces, e.g. (always_on | always_off | traceidratio | parentbased_always_on | parentbased_always_off | parentbased_traceidratio | parentbased_jaeger_remote | jaeger_remote | xray)
	OtelTracesSampler string

	// String value to be used as the sampler argument
	OtelTracesSamplerArg string

	// Trace exporter to be used (otlp | zipkin | none)
	OtelTracesExporter string

	// Metrics exporter to be used (otlp | prometheus | none)
	OtelMetricsExporter string

	// Logs exporter to be used (otlp | none)
	OtelLogsExporter string
}

func (config *Config) GetConfigFlagSet(fs *flag.FlagSet) {
	// No flagset item is added, only the environment values are used with defaults
	config.ServiceName = env.GetEnvWithDefault(serviceNameEnvVar, ServiceNameDefault)
	config.OtelTracesSampler = env.GetEnvWithDefault(otelTracesSamplerEnvVar, OtelTracesSamplerDefault)
	config.OtelTracesSamplerArg = env.GetEnvWithDefault(otelTracesSamplerArgEnvVar, OtelTracesSamplerArgDefault)
	config.OtelTracesExporter = env.GetEnvWithDefault(otelTracesExporterEnvVar, OtelTracesExporterDefault)
	config.OtelMetricsExporter = env.GetEnvWithDefault(otelMetricsExporterEnvVar, OtelMetricsExporterDefault)
	config.OtelLogsExporter = env.GetEnvWithDefault(otelLogsExporterEnvVar, OtelLogsExporterDefault)
}
