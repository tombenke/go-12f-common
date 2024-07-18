package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/examples/scheduler/scheduler"
	"github.com/tombenke/go-12f-common/examples/scheduler/timer"
	"go.uber.org/multierr"
)

// The configuration parameters of the Application
type Config struct {
	scheduler scheduler.Config
	timer     timer.Config
}

func (c *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	c.scheduler.GetConfigFlagSet(flagSet)
	c.timer.GetConfigFlagSet(flagSet)
}

func (c *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := apprun.LoadConfigWithDefaultViper(flagSet, c); err != nil {
		return fmt.Errorf("failed to load config. %w", err)
	}
	return multierr.Combine(
		c.timer.LoadConfig(flagSet),
		c.scheduler.LoadConfig(flagSet),
	)
}

var _ apprun.Configurer = (*Config)(nil)
