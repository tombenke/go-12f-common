package otel_test

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/otel"
	"testing"
)

func TestOtelConfigWithDefaults(t *testing.T) {
	config := otel.Config{}
	fs := flag.NewFlagSet("test-fs", flag.ContinueOnError)
	config.GetConfigFlagSet(fs)
	assert.Equal(t, otel.Config{
		ServiceName:          otel.ServiceNameDefault,
		OtelTracesSampler:    otel.OtelTracesSamplerDefault,
		OtelTracesSamplerArg: otel.OtelTracesSamplerArgDefault,
		OtelTracesExporter:   otel.OtelTracesExporterDefault,
		OtelMetricsExporter:  otel.OtelMetricsExporterDefault,
		OtelLogsExporter:     otel.OtelLogsExporterDefault,
	}, config)
}
