package oti

import (
	"fmt"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOtelConfigWithDefaults(t *testing.T) {
	config := Config{}
	fs := pflag.NewFlagSet("test-fs", pflag.ContinueOnError)
	config.GetConfigFlagSet(fs)
	require.NoError(t, config.LoadConfig(fs))
	assert.Equal(t, Config{
		OtelTracesExporter:         OTEL_TRACES_EXPORTER_DEFAULT,
		OtelMetricsExporter:        OTEL_METRICS_EXPORTER_DEFAULT,
		OtelExporterPrometheusPort: OTEL_EXPORTER_PROMETHEUS_PORT_DEFAULT,
	}, config)
}

func Test_Config_GetConfigFlagSet(t *testing.T) {
	const EXPECTED_OTEL_TRACES_EXPORTER_FROM_ENV_VAR = "env_jaeger"
	const EXPECTED_OTEL_TRACES_EXPORTER_FROM_CLI_ARG = "cli_jaeger"

	const EXPECTED_OTEL_METRICS_EXPORTER_FROM_ENV_VAR = "env_prometheus"
	const EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG = "cli_prometheus"

	const EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_ENV_VAR = 1234
	const EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_CLI_ARG = 5678

	envVars := map[string]string{
		"OTEL_TRACES_EXPORTER":          EXPECTED_OTEL_TRACES_EXPORTER_FROM_ENV_VAR,
		"OTEL_METRICS_EXPORTER":         EXPECTED_OTEL_METRICS_EXPORTER_FROM_ENV_VAR,
		"OTEL_EXPORTER_PROMETHEUS_PORT": fmt.Sprintf("%v", EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_ENV_VAR),
	}
	cliArgs := []string{
		fmt.Sprintf("--%v=%v", OTEL_TRACES_EXPORTER_ARG_NAME, EXPECTED_OTEL_TRACES_EXPORTER_FROM_CLI_ARG),
		fmt.Sprintf("--%v=%v", OTEL_METRICS_EXPORTER_ARG_NAME, EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG),
		fmt.Sprintf("--%v=%v", OTEL_EXPORTER_PROMETHEUS_PORT_ARG_NAME, EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_CLI_ARG),
	}
	testCases := map[string]struct {
		expectedConfig Config
		envVars        map[string]string
		cliArgs        []string
	}{
		"default values": {
			expectedConfig: Config{
				OtelTracesExporter:         OTEL_TRACES_EXPORTER_DEFAULT,
				OtelMetricsExporter:        OTEL_METRICS_EXPORTER_DEFAULT,
				OtelExporterPrometheusPort: OTEL_EXPORTER_PROMETHEUS_PORT_DEFAULT,
			},
		},
		"from environment variables": {
			expectedConfig: Config{
				OtelTracesExporter:         EXPECTED_OTEL_TRACES_EXPORTER_FROM_ENV_VAR,
				OtelMetricsExporter:        EXPECTED_OTEL_METRICS_EXPORTER_FROM_ENV_VAR,
				OtelExporterPrometheusPort: EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_ENV_VAR,
			},
			envVars: envVars,
		},
		"from cli args": {
			expectedConfig: Config{
				OtelTracesExporter:         EXPECTED_OTEL_TRACES_EXPORTER_FROM_CLI_ARG,
				OtelMetricsExporter:        EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG,
				OtelExporterPrometheusPort: EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_CLI_ARG,
			},
			cliArgs: cliArgs,
		},
		"prefer cli args over env vars": {
			expectedConfig: Config{
				OtelTracesExporter:         EXPECTED_OTEL_TRACES_EXPORTER_FROM_CLI_ARG,
				OtelMetricsExporter:        EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG,
				OtelExporterPrometheusPort: EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_CLI_ARG,
			},
			envVars: envVars,
			cliArgs: cliArgs,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			// given
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			cfg := &Config{}

			for k, v := range testCase.envVars {
				t.Setenv(k, v)
			}

			// when
			cfg.GetConfigFlagSet(fs)
			require.NoError(t, fs.Parse(testCase.cliArgs))

			err := cfg.LoadConfig(fs)

			// then
			assert := assert.New(t)
			assert.NoError(err)
			assert.Equal(testCase.expectedConfig, *cfg)
		})
	}
}
