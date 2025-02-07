package otel

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/apprun"
)

const (
	OTEL_SERVICE_NAME_ARG_NAME = "otel-service-name"
	OTEL_SERVICE_NAME_DEFAULT  = "undefined" // TODO: Its default value shoul come from the appname provided by the buildinfo
	OTEL_SERVICE_NAME_HELP     = "ServiceName is the name of the service that collects metrics and tracing"

// //	ConfigServiceNameDefault      = "undefined"
// //	ConfigTracesSamplerDefault    = "always_off"
// //	ConfigTracesSamplerArgDefault = ""
// //	ConfigTracesExporterDefault   = "otlp"
// //	ConfigMetricsExporterDefault  = "otlp"
// //	ConfigLogsExporterDefault     = "otlp"
)

// Config holds the configuration parameters for the Open Telemetry instrumentation
type Config struct {
	// ServiceName is the name of the service that collects metrics and tracing
	OtelServiceName string `mapstructure:"otel-service-name"`

	// OtelResourceAttributes is the environment variable name OpenTelemetry Resource information will be read from
	// in the form of commaseparated key-value pairs.
	OtelResourceAttributes string `mapstructure:"otel-resource-attributes"`

	// OtelMetricsExporter specifies which exporter is used for metrics.
	// Possible values are: "otlp": OTLP, "prometheus": Prometheus, "console": Standard Output, "none": No automatically configured exporter for metrics
	OtelMetricsExporter string `mapstructure:"otel-metrics-exporter"`

	// //	// Sampler to be used for traces, e.g. (always_on | always_off | traceidratio | parentbased_always_on | parentbased_always_off | parentbased_traceidratio | parentbased_jaeger_remote | jaeger_remote | xray)
	// //	TracesSampler string `mapstructure:"traces-sampler"`
	// //	// String value to be used as the sampler argument
	// //	TracesSamplerArg string `mapstructure:"traces-sampler-arg"`
	// //	// Trace exporter to be used (otlp | zipkin | none)
	// //	TracesExporter string `mapstructure:"traces-exporter"`
	// //	// Metrics exporter to be used (otlp | prometheus | none)
	// //	MetricsExporter string `mapstructure:"metrics-exporter"`
	// //	// Logs exporter to be used (otlp | none)
	// //	LogsExporter string `mapstructure:"logs-exporter"`
}

func (cfg *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	flagSet.String(OTEL_SERVICE_NAME_ARG_NAME, OTEL_SERVICE_NAME_DEFAULT, OTEL_SERVICE_NAME_HELP)
}

func (cfg *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := apprun.LoadConfigWithDefaultViper(flagSet, cfg); err != nil {
		return fmt.Errorf("failed to load otel config. %w", err)
	}
	return nil
}

////var _ apprun.Configurer = (*Config)(nil)
