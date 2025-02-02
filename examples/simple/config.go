package main

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/apprun"
)

const (
	STARTUP_DELAY_ARG_NAME = "startup-delay"
	STARTUP_DELAY_DEFAULT  = time.Duration(30 * time.Second)
	STARTUP_DELAY_HELP     = "The delay of startup process"
)

// The configuration parameters of the Application
type Config struct {
	StartupDelay time.Duration `mapstructure:"startup-delay"`
}

func (c *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	flagSet.Duration(STARTUP_DELAY_ARG_NAME, STARTUP_DELAY_DEFAULT, STARTUP_DELAY_HELP)
}

func (c *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := apprun.LoadConfigWithDefaultViper(flagSet, c, "simple"); err != nil {
		return fmt.Errorf("failed to load otel config. %w", err)
	}
	return nil
}

var _ apprun.Configurer = (*Config)(nil)
