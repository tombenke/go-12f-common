package main

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/apprun"
)

// The configuration parameters of the Application
type Config struct {
	StartupDelay time.Duration `mapstructure:"startup-delay"`
}

func (c *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	flagSet.Duration("startup-delay", 30*time.Second, "The delay of startup process")
}

func (c *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := apprun.LoadConfigWithDefaultViper(flagSet, c); err != nil {
		return fmt.Errorf("failed to load otel config. %w", err)
	}
	return nil
}

var _ apprun.Configurer = (*Config)(nil)
