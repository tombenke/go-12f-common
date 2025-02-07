package apprun

import (
	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/config"
	"github.com/tombenke/go-12f-common/oti"
)

const (
	LogLevelDefault  = "info"
	LogFormatDefault = "json"

	HealthCheckPortDefault    = 8080
	LivenessCheckPathDefault  = "/live"
	ReadinessCheckPathDefault = "/ready"
)

// Config represents the main configuration object of the 12-factor application instance
// It holds those parameters that are needed to setup the basic functionalities of the application,
// e.g. logging, healthcheck, levness and readiness checks.
type Config struct {
	LogLevel           string `mapstructure:"log-level"`
	LogFormat          string `mapstructure:"log-format"`
	HealthCheckPort    uint   `mapstructure:"health-check-port"`
	LivenessCheckPath  string `mapstructure:"liveness-check-path"`
	ReadinessCheckPath string `mapstructure:"readiness-check-path"`
	OtelConfig         oti.Config
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
	if err := config.LoadConfigWithDefaultViper(flagSet, cfg); err != nil {
		return err
	}
	return cfg.OtelConfig.LoadConfig(flagSet)
}

// Ensure that Config implements the Configurer interface
var _ config.Configurer = (*Config)(nil)
