package otel_test

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tombenke/go-12f-common/otel"
)

func TestOtelConfigWithDefaults(t *testing.T) {
	config := otel.Config{}
	fs := pflag.NewFlagSet("test-fs", pflag.ContinueOnError)
	config.GetConfigFlagSet(fs)
	require.NoError(t, config.LoadConfig(fs))
	assert.Equal(t, otel.Config{
		OtelServiceName: otel.OTEL_SERVICE_NAME_DEFAULT,
		////		TracesSampler:    otel.ConfigTracesSamplerDefault,
		////		TracesSamplerArg: otel.ConfigTracesSamplerArgDefault,
		////		TracesExporter:   otel.ConfigTracesExporterDefault,
		////		MetricsExporter:  otel.ConfigMetricsExporterDefault,
		////		LogsExporter:     otel.ConfigLogsExporterDefault,
	}, config)
}
