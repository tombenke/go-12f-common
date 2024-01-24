package main

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tombenke/go-12f-common/apprun"
)

const (
	ServerPortHelp    = "The HTTP port of the server"
	ServerPortDefault = 8081
)

// The configuration parameters of the ExampleApp
type Config struct {
	ServerPort uint `mapstructure:"server-port"`
}

func (c *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	flagSet.Uint("server-port", ServerPortDefault, ServerPortHelp)

}

func (c *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	viper := viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer("-", "_")))
	if err := viper.BindPFlags(flagSet); err != nil {
		return fmt.Errorf("failed to bind flag set to config. %w", err)
	}
	viper.AutomaticEnv()

	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarshal into simple application config. %w", err)
	}
	return nil
}

var _ apprun.Configurer = (*Config)(nil)
