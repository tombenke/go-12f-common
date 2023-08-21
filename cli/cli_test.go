package cli_test

import (
	"flag"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/cli"
	"github.com/tombenke/go-12f-common/env"
	"testing"
)

const (
	boolVarHelp    = "BoolVar's help text"
	boolVarDefault = true

	stringVarEnvVar  = "STRING_VAR"
	stringVarDefault = "StringVar's default value"
	stringVarHelp    = "StringVar help text"

	intVarEnvVar  = "INT_VAR"
	intVarDefault = "42"
	intVarHelp    = "IntVar's help text"
)

type TestConfig struct {
	BoolVar   bool
	StringVar string
	IntVar    int
}

func (tc *TestConfig) GetConfigFlagSet(fs *flag.FlagSet) {
	fs.BoolVar(&tc.BoolVar, "b", boolVarDefault, boolVarHelp)
	fs.StringVar(&tc.StringVar, "s", env.GetEnvWithDefault(stringVarEnvVar, stringVarDefault), stringVarHelp)
	fs.IntVar(&tc.IntVar, "i", int(env.GetEnvWithDefaultUint(intVarEnvVar, intVarDefault)), intVarHelp)
}

func TestInitConfigs(t *testing.T) {
	testArgs := []string{"test-app"}

	testConfig := TestConfig{}
	cli.InitConfigs(testArgs, []cli.FlagSetFunc{
		testConfig.GetConfigFlagSet,
	})
	fmt.Printf("%v", testConfig)
	assert.Equal(t, TestConfig{BoolVar: boolVarDefault, StringVar: stringVarDefault, IntVar: int(42)}, testConfig)
}
