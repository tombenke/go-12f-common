package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/config"
	"github.com/tombenke/go-12f-common/examples/scheduler/timer"
	"github.com/tombenke/go-12f-common/examples/scheduler/worker"
	"go.uber.org/multierr"
)

// The configuration parameters of the Application
type Config struct {
	worker worker.Config
	timer  timer.Config
}

func (c *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	c.worker.GetConfigFlagSet(flagSet)
	c.timer.GetConfigFlagSet(flagSet)
}

func (c *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := config.LoadConfigWithDefaultViper(flagSet, c); err != nil {
		return fmt.Errorf("failed to load config. %w", err)
	}
	return multierr.Combine(
		c.timer.LoadConfig(flagSet),
		c.worker.LoadConfig(flagSet),
	)
}

var _ config.Configurer = (*Config)(nil)
