package oti_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tombenke/go-12f-common/oti"
)

func TestOtelConfigWithDefaults(t *testing.T) {
	config := oti.Config{}
	fs := pflag.NewFlagSet("test-fs", pflag.ContinueOnError)
	config.GetConfigFlagSet(fs)
	require.NoError(t, config.LoadConfig(fs))
	assert.Equal(t, oti.Config{
		OtelServiceName:            oti.OTEL_SERVICE_NAME_DEFAULT,
		OtelResourceAttributes:     oti.OTEL_RESOURCE_ATTRIBUTES_DEFAULT,
		OtelMetricsExporter:        oti.OTEL_METRICS_EXPORTER_DEFAULT,
		OtelExporterPrometheusPort: oti.OTEL_EXPORTER_PROMETHEUS_PORT_DEFAULT,
	}, config)
}

func Test_Config_GetConfigFlagSet(t *testing.T) {
	const EXPECTED_OTEL_SERVICE_NAME_FROM_ENV_VAR = "env_servicename"
	const EXPECTED_OTEL_SERVICE_NAME_FROM_CLI_ARG = "cli_servicename"

	const EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_ENV_VAR = "env_attr1=value1,attr2=value2"
	const EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_CLI_ARG = "cli_attr1=value1,attr2=value2"

	const EXPECTED_OTEL_METRICS_EXPORTER_FROM_ENV_VAR = "env_prometheus"
	const EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG = "cli_prometheus"

	const EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_ENV_VAR = 1234
	const EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_CLI_ARG = 5678

	envVars := map[string]string{
		"OTEL_SERVICE_NAME":             EXPECTED_OTEL_SERVICE_NAME_FROM_ENV_VAR,
		"OTEL_RESOURCE_ATTRIBUTES":      EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_ENV_VAR,
		"OTEL_METRICS_EXPORTER":         EXPECTED_OTEL_METRICS_EXPORTER_FROM_ENV_VAR,
		"OTEL_EXPORTER_PROMETHEUS_PORT": fmt.Sprintf("%v", EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_ENV_VAR),
	}
	cliArgs := []string{
		fmt.Sprintf("--%v=%v", oti.OTEL_SERVICE_NAME_ARG_NAME, EXPECTED_OTEL_SERVICE_NAME_FROM_CLI_ARG),
		fmt.Sprintf("--%v=%v", oti.OTEL_RESOURCE_ATTRIBUTES_ARG_NAME, EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_CLI_ARG),
		fmt.Sprintf("--%v=%v", oti.OTEL_METRICS_EXPORTER_ARG_NAME, EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG),
		fmt.Sprintf("--%v=%v", oti.OTEL_EXPORTER_PROMETHEUS_PORT_ARG_NAME, EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_CLI_ARG),
	}
	testCases := map[string]struct {
		expectedConfig oti.Config
		envVars        map[string]string
		cliArgs        []string
	}{
		"default values": {
			expectedConfig: oti.Config{
				OtelServiceName:            oti.OTEL_SERVICE_NAME_DEFAULT,
				OtelResourceAttributes:     oti.OTEL_RESOURCE_ATTRIBUTES_DEFAULT,
				OtelMetricsExporter:        oti.OTEL_METRICS_EXPORTER_DEFAULT,
				OtelExporterPrometheusPort: oti.OTEL_EXPORTER_PROMETHEUS_PORT_DEFAULT,
			},
		},
		"from environment variables": {
			expectedConfig: oti.Config{
				OtelServiceName:            EXPECTED_OTEL_SERVICE_NAME_FROM_ENV_VAR,
				OtelResourceAttributes:     EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_ENV_VAR,
				OtelMetricsExporter:        EXPECTED_OTEL_METRICS_EXPORTER_FROM_ENV_VAR,
				OtelExporterPrometheusPort: EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_ENV_VAR,
			},
			envVars: envVars,
		},
		"from cli args": {
			expectedConfig: oti.Config{
				OtelServiceName:            EXPECTED_OTEL_SERVICE_NAME_FROM_CLI_ARG,
				OtelResourceAttributes:     EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_CLI_ARG,
				OtelMetricsExporter:        EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG,
				OtelExporterPrometheusPort: EXPECTED_OTEL_EXPORTER_PROMETHEUS_PORT_FROM_CLI_ARG,
			},
			cliArgs: cliArgs,
		},
		"prefer cli args over env vars": {
			expectedConfig: oti.Config{
				OtelServiceName:            EXPECTED_OTEL_SERVICE_NAME_FROM_CLI_ARG,
				OtelResourceAttributes:     EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_CLI_ARG,
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
			cfg := &oti.Config{}

			for k, v := range testCase.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
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
