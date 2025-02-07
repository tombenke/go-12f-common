package otel_test

import (
	"fmt"
	"os"
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
		OtelServiceName:        otel.OTEL_SERVICE_NAME_DEFAULT,
		OtelResourceAttributes: otel.OTEL_RESOURCE_ATTRIBUTES_DEFAULT,
		OtelMetricsExporter:    otel.OTEL_METRICS_EXPORTER_DEFAULT,
	}, config)
}

func Test_Config_GetConfigFlagSet(t *testing.T) {
	const EXPECTED_OTEL_SERVICE_NAME_FROM_ENV_VAR = "env_servicename"
	const EXPECTED_OTEL_SERVICE_NAME_FROM_CLI_ARG = "cli_servicename"

	const EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_ENV_VAR = "env_attr1=value1,attr2=value2"
	const EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_CLI_ARG = "cli_attr1=value1,attr2=value2"

	const EXPECTED_OTEL_METRICS_EXPORTER_FROM_ENV_VAR = "env_prometheus"
	const EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG = "cli_prometheus"

	envVars := map[string]string{
		"OTEL_SERVICE_NAME":        EXPECTED_OTEL_SERVICE_NAME_FROM_ENV_VAR,
		"OTEL_RESOURCE_ATTRIBUTES": EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_ENV_VAR,
		"OTEL_METRICS_EXPORTER":    EXPECTED_OTEL_METRICS_EXPORTER_FROM_ENV_VAR,
	}
	cliArgs := []string{
		fmt.Sprintf("--%v=%v", otel.OTEL_SERVICE_NAME_ARG_NAME, EXPECTED_OTEL_SERVICE_NAME_FROM_CLI_ARG),
		fmt.Sprintf("--%v=%v", otel.OTEL_RESOURCE_ATTRIBUTES_ARG_NAME, EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_CLI_ARG),
		fmt.Sprintf("--%v=%v", otel.OTEL_METRICS_EXPORTER_ARG_NAME, EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG),
	}
	testCases := map[string]struct {
		expectedConfig otel.Config
		envVars        map[string]string
		cliArgs        []string
	}{
		"default values": {
			expectedConfig: otel.Config{
				OtelServiceName:        otel.OTEL_SERVICE_NAME_DEFAULT,
				OtelResourceAttributes: otel.OTEL_RESOURCE_ATTRIBUTES_DEFAULT,
				OtelMetricsExporter:    otel.OTEL_METRICS_EXPORTER_DEFAULT,
			},
		},
		"from environment variables": {
			expectedConfig: otel.Config{
				OtelServiceName:        EXPECTED_OTEL_SERVICE_NAME_FROM_ENV_VAR,
				OtelResourceAttributes: EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_ENV_VAR,
				OtelMetricsExporter:    EXPECTED_OTEL_METRICS_EXPORTER_FROM_ENV_VAR,
			},
			envVars: envVars,
		},
		"from cli args": {
			expectedConfig: otel.Config{
				OtelServiceName:        EXPECTED_OTEL_SERVICE_NAME_FROM_CLI_ARG,
				OtelResourceAttributes: EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_CLI_ARG,
				OtelMetricsExporter:    EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG,
			},
			cliArgs: cliArgs,
		},
		"prefer cli args over env vars": {
			expectedConfig: otel.Config{
				OtelServiceName:        EXPECTED_OTEL_SERVICE_NAME_FROM_CLI_ARG,
				OtelResourceAttributes: EXPECTED_OTEL_RESOURCE_ATTRIBUTES_FROM_CLI_ARG,
				OtelMetricsExporter:    EXPECTED_OTEL_METRICS_EXPORTER_FROM_CLI_ARG,
			},
			envVars: envVars,
			cliArgs: cliArgs,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			// given
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			cfg := &otel.Config{}

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
