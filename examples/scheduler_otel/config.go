package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"go.uber.org/multierr"

	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/examples/scheduler_otel/buildinfo"
	"github.com/tombenke/go-12f-common/examples/scheduler_otel/timer"
	"github.com/tombenke/go-12f-common/examples/scheduler_otel/worker"
)

// The configuration parameters of the Application
type Config struct {
	QueueSize int `mapstructure:"queue-size"`

	worker worker.Config
	timer  timer.Config
}

func (c *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	c.worker.GetConfigFlagSet(flagSet)
	c.timer.GetConfigFlagSet(flagSet)
}

func (c *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := apprun.LoadConfigWithDefaultViper(flagSet, c, buildinfo.BuildInfo.AppName()); err != nil {
		return fmt.Errorf("failed to load config. %w", err)
	}
	return multierr.Combine(
		c.timer.LoadConfig(flagSet),
		c.worker.LoadConfig(flagSet),
	)
}

var _ apprun.Configurer = (*Config)(nil)
