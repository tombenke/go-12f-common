package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/apprun"
)

// The configuration parameters of the ExampleApp
type Config struct {
	ServerPort uint `mapstructure:"server-port"`
}

func (c *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	flagSet.Uint("server-port", 8081, "The HTTP port of the server")

}

func (c *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := apprun.LoadConfigWithDefaultViper(flagSet, c, "otel"); err != nil {
		return fmt.Errorf("failed to load otel config. %w", err)
	}
	return nil
}

var _ apprun.Configurer = (*Config)(nil)
