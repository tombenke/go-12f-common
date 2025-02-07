package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Configurer defines the interface for the application and its components that needs an kind of configurability
type Configurer interface {
	// GetConfigFlagSet is a factory function that receives a reference to a `pflag.FlagSet` object, to that it puts its configuration parameters
	GetConfigFlagSet(flagSet *pflag.FlagSet)

	// LoadConfig resolves the actual values of the configuration object.
	// It takes into account the parameter definitions, the CLI and environment variables and the default values as well.
	LoadConfig(flagSet *pflag.FlagSet) error
}

func NewDefaultViper(flagSet *pflag.FlagSet) (*viper.Viper, error) {
	viper := viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer("-", "_")))
	if err := viper.BindPFlags(flagSet); err != nil {
		return nil, fmt.Errorf("failed to bind flag set to config. %w", err)
	}
	viper.AutomaticEnv()
	return viper, nil
}

func LoadConfigWithDefaultViper(flagSet *pflag.FlagSet, config any) error {
	if reflect.ValueOf(config).Kind() != reflect.Ptr {
		panic("config must be a pointer")
	}
	viper, err := NewDefaultViper(flagSet)
	if err != nil {
		return err
	}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal into config. %w", err)
	}
	return nil
}
