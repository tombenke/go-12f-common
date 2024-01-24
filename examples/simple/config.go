package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tombenke/go-12f-common/apprun"
)

const (
	StartupDelayHelp    = "The delay of startup process"
	StartupDelayDefault = 30 * time.Second
)

// The configuration parameters of the Application
type Config struct {
	StartupDelay time.Duration `mapstructure:"startup-delay"`
}

func (c *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	flagSet.Duration("startup-delay", StartupDelayDefault, StartupDelayHelp)
}

func (c *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	viper := viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer("-", "_")))
	if err := viper.BindPFlags(flagSet); err != nil {
		return fmt.Errorf("failed to bind flag set to simple application config. %w", err)
	}
	viper.AutomaticEnv()

	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarshal into simple application config. %w", err)
	}
	return nil
}

var _ apprun.Configurer = (*Config)(nil)
