package timer

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Config with default values
func TestConfigDefaults(t *testing.T) {
	// TODO
	config := Config{}

	fs := &pflag.FlagSet{}
	config.GetConfigFlagSet(fs)
	fmt.Printf("%+v, %+v\n", config, *fs)
	config.LoadConfig(fs)
	fmt.Printf("%+v, %+v\n", config, *fs)
	assert.Equal(t, TIME_STEP_DEFAULT, config.TimeStep)
}

// Config with environment variables
func TestConfigWithEnvironment(t *testing.T) {
	// TODO
	assert.Nil(t, nil)
}

// Config with command-line arguments
func TestConfigWithCliArgs(t *testing.T) {
	// TODO
	assert.Nil(t, nil)
}
