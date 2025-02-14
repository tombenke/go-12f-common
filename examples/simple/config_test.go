package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tombenke/go-12f-common/v2/must"
)

func Test_Config_GetConfigFlagSet(t *testing.T) {
	const EXPECTED_STARTUP_DELAY_FROM_ENV_VAR = "10s"
	const EXPECTED_STARTUP_DELAY_FROM_CLI_ARG = "20s"

	envVars := map[string]string{
		"STARTUP_DELAY": EXPECTED_STARTUP_DELAY_FROM_ENV_VAR,
	}
	cliArgs := []string{
		fmt.Sprintf("--%v=%v", STARTUP_DELAY_ARG_NAME, EXPECTED_STARTUP_DELAY_FROM_CLI_ARG),
	}
	testCases := map[string]struct {
		expectedConfig Config
		envVars        map[string]string
		cliArgs        []string
	}{
		"default values": {
			expectedConfig: Config{
				StartupDelay: STARTUP_DELAY_DEFAULT,
			},
		},
		"from environment variables": {
			expectedConfig: Config{
				StartupDelay: must.MustVal(time.ParseDuration(EXPECTED_STARTUP_DELAY_FROM_ENV_VAR)),
			},
			envVars: envVars,
		},
		"from cli args": {
			expectedConfig: Config{
				StartupDelay: must.MustVal(time.ParseDuration(EXPECTED_STARTUP_DELAY_FROM_CLI_ARG)),
			},
			cliArgs: cliArgs,
		},
		"prefer cli args over env vars": {
			expectedConfig: Config{
				StartupDelay: must.MustVal(time.ParseDuration(EXPECTED_STARTUP_DELAY_FROM_CLI_ARG)),
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
