package oti

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/config"
)

const (
	OTEL_TRACES_EXPORTER_ARG_NAME = "otel-traces-exporter"
	OTEL_TRACES_EXPORTER_DEFAULT  = "none"
	OTEL_TRACES_EXPORTER_HELP     = "Selects the exporter to use for tracing: otlp | console | none"

	OTEL_METRICS_EXPORTER_ARG_NAME = "otel-metrics-exporter"
	OTEL_METRICS_EXPORTER_DEFAULT  = "none"
	OTEL_METRICS_EXPORTER_HELP     = "Selects the exporter to use for metrics: otlp | prometheus | console | none"

	OTEL_EXPORTER_PROMETHEUS_PORT_ARG_NAME = "otel-exporter-prometheus-port"
	OTEL_EXPORTER_PROMETHEUS_PORT_DEFAULT  = 9464
	OTEL_EXPORTER_PROMETHEUS_PORT_HELP     = "the port used by the Prometheus exporter"
)

// Config holds the configuration parameters for the Open Telemetry instrumentation
type Config struct {
	// OtelTracesExporter specifies which exporter is used for tracing
	// Possible values are: "otlp": OTLP, "jaeger": Jaeger, "zipkin": Zipkin, "console": Standard Output, "none": No automatically configured exporter for tracing
	OtelTracesExporter string `mapstructure:"otel-traces-exporter"`

	// OtelMetricsExporter specifies which exporter is used for metrics
	// Possible values are: "otlp": OTLP, "prometheus": Prometheus, "console": Standard Output, "none": No automatically configured exporter for metrics
	OtelMetricsExporter string `mapstructure:"otel-metrics-exporter"`

	// OtelExporterPrometheusPort specifies the port that the prometheus exporter uses to provide the metrics
	OtelExporterPrometheusPort int `mapstructure:"otel-exporter-prometheus-port"`
}

func (cfg *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	flagSet.String(OTEL_TRACES_EXPORTER_ARG_NAME, OTEL_TRACES_EXPORTER_DEFAULT, OTEL_TRACES_EXPORTER_HELP)
	flagSet.String(OTEL_METRICS_EXPORTER_ARG_NAME, OTEL_METRICS_EXPORTER_DEFAULT, OTEL_METRICS_EXPORTER_HELP)
	flagSet.Int(OTEL_EXPORTER_PROMETHEUS_PORT_ARG_NAME, OTEL_EXPORTER_PROMETHEUS_PORT_DEFAULT, OTEL_EXPORTER_PROMETHEUS_PORT_HELP)
}

func (cfg *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := config.LoadConfigWithDefaultViper(flagSet, cfg); err != nil {
		return fmt.Errorf("failed to load otel config. %w", err)
	}
	return nil
}

var _ config.Configurer = (*Config)(nil)
