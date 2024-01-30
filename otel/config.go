package otel

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ConfigServiceNameDefault      = "undefined"
	ConfigTracesSamplerDefault    = "always_off"
	ConfigTracesSamplerArgDefault = ""
	ConfigTracesExporterDefault   = "otlp"
	ConfigMetricsExporterDefault  = "otlp"
	ConfigLogsExporterDefault     = "otlp"
)

// Config holds the configuration parameters for the Open Telemetry instrumentation
type Config struct {
	// ServiceName is the name of the service that collects metrics and tracing
	ServiceName string `mapstructure:"service-name"`
	// Sampler to be used for traces, e.g. (always_on | always_off | traceidratio | parentbased_always_on | parentbased_always_off | parentbased_traceidratio | parentbased_jaeger_remote | jaeger_remote | xray)
	TracesSampler string `mapstructure:"traces-sampler"`
	// String value to be used as the sampler argument
	TracesSamplerArg string `mapstructure:"traces-sampler-arg"`
	// Trace exporter to be used (otlp | zipkin | none)
	TracesExporter string `mapstructure:"traces-exporter"`
	// Metrics exporter to be used (otlp | prometheus | none)
	MetricsExporter string `mapstructure:"metrics-exporter"`
	// Logs exporter to be used (otlp | none)
	LogsExporter string `mapstructure:"logs-exporter"`
}

func (cfg *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
}

func (cfg *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	viper := viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer("-", "_")))
	viper.SetEnvPrefix("otel")
	viper.SetDefault("service-name", ConfigServiceNameDefault)
	viper.SetDefault("traces-sampler", ConfigTracesSamplerDefault)
	viper.SetDefault("traces-sampler-arg", ConfigTracesSamplerArgDefault)
	viper.SetDefault("traces-exporter", ConfigTracesExporterDefault)
	viper.SetDefault("metrics-exporter", ConfigMetricsExporterDefault)
	viper.SetDefault("logs-exporter", ConfigLogsExporterDefault)
	viper.AutomaticEnv()

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal otel config. %w", err)
	}
	return nil
}
