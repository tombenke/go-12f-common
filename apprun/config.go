package apprun

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tombenke/go-12f-common/otel"
)

const (
	LogLevelDefault  = "info"
	LogFormatDefault = "json"

	HealthCheckPortDefault    = 8080
	LivenessCheckPathDefault  = "/live"
	ReadinessCheckPathDefault = "/ready"
)

// Configurer defines the interface for the application and its components that needs an kind of configurability
type Configurer interface {
	// GetConfigFlagSet is a factory function that receives a reference to a `pflag.FlagSet` object, to that it puts its configuration parameters
	GetConfigFlagSet(flagSet *pflag.FlagSet)

	// LoadConfig resolves the actual values of the configuration object.
	// It takes into account the parameter definitions, the CLI and environment variables and the default values as well.
	LoadConfig(flagSet *pflag.FlagSet) error
}

// Config represents the main configuration object of the 12-factor application instance
// It holds those parameters that are needed to setup the basic functionalities of the application,
// e.g. logging, healthcheck, levness and readiness checks.
type Config struct {
	LogLevel           string `mapstructure:"log-level"`
	LogFormat          string `mapstructure:"log-format"`
	HealthCheckPort    uint   `mapstructure:"health-check-port"`
	LivenessCheckPath  string `mapstructure:"liveness-check-path"`
	ReadinessCheckPath string `mapstructure:"readiness-check-path"`
	OtelConfig         otel.Config
}

// GetConfigFlagSet() initializes the configuration object of the 12-factor application, and returns with it
func (cfg *Config) GetConfigFlagSet(flagSet *pflag.FlagSet) {
	flagSet.StringP(
		"log-level",
		"l",
		LogLevelDefault,
		"The log level: panic | fatal | error | warning | info | debug | trace",
	)
	flagSet.StringP("log-format", "f", LogFormatDefault, "The log format: json | text")

	// HealthCheck parameters
	flagSet.Uint("health-check-port", HealthCheckPortDefault, "The HTTP port of the healthcheck endpoints")
	flagSet.String("liveness-check-path", LivenessCheckPathDefault, "The path of the liveness check endpoint")
	flagSet.String("readiness-check-path", ReadinessCheckPathDefault, "The path of the readiness check endpoint")

	cfg.OtelConfig.GetConfigFlagSet(flagSet)
}

func (cfg *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := LoadConfigWithDefaultViper(flagSet, cfg, "apprun"); err != nil {
		return err
	}
	return cfg.OtelConfig.LoadConfig(flagSet)
}

func NewDefaultViper(flagSet *pflag.FlagSet, app string) (*viper.Viper, error) {
	viper := viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer("-", "_")))
	if err := viper.BindPFlags(flagSet); err != nil {
		return nil, fmt.Errorf("failed to bind flag set to config. %w", err)
	}
	viper.AutomaticEnv()
	viper.SetConfigName(app)    // name of config file (without extension)
	viper.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")
	viper.ReadInConfig()
	return viper, nil
}

func LoadConfigWithDefaultViper(flagSet *pflag.FlagSet, config any, app string) error {
	if reflect.ValueOf(config).Kind() != reflect.Ptr {
		panic("config must be a pointer")
	}
	viper, err := NewDefaultViper(flagSet, app)
	if err != nil {
		return err
	}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal into config. %w", err)
	}
	return nil
}

// Ensure that Config implements the Configurer interface
var _ Configurer = (*Config)(nil)
