package timer

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/apprun"
)

const (
	TIME_STEP_HELP     = "The size of a time-step"
	TIME_STEP_ARG_NAME = "time-step"
	TIME_STEP_DEFAULT  = "60s"
)

// The configuration parameters of the Timer component
type Config struct {
	TimeStep string `mapstructure:"time-step"`
}

// Add application-specific config parameters to flagset
func (cfg *Config) GetConfigFlagSet(fs *pflag.FlagSet) {
	fs.String(TIME_STEP_ARG_NAME, TIME_STEP_DEFAULT, TIME_STEP_HELP)
}

func (cfg *Config) LoadConfig(fs *pflag.FlagSet) error {
	if err := apprun.LoadConfigWithDefaultViper(fs, cfg); err != nil {
		return fmt.Errorf("failed to load config. %w", err)
	}
	return nil
}

var _ apprun.Configurer = (*Config)(nil)
