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

type Configurer interface {
	GetConfigFlagSet(flagSet *pflag.FlagSet)
	LoadConfig(flagSet *pflag.FlagSet) error
}

type Config struct {
	LogLevel           string `mapstructure:"log-level"`
	LogFormat          string `mapstructure:"log-format"`
	HealthCheckPort    uint   `mapstructure:"health-check-port"`
	LivenessCheckPath  string `mapstructure:"liveness-check-path"`
	ReadinessCheckPath string `mapstructure:"readiness-check-path"`
	OtelConfig         otel.Config
}

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
	if err := LoadConfigWithDefaultViper(flagSet, cfg); err != nil {
		return err
	}
	return cfg.OtelConfig.LoadConfig(flagSet)
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

var _ Configurer = (*Config)(nil)
