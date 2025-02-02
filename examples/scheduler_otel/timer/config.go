package timer

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/examples/scheduler_otel/buildinfo"
)

const (
	TIME_STEP_HELP     = "The size of a time-step"
	TIME_STEP_ARG_NAME = "time-step"
	TIME_STEP_DEFAULT  = "60s"
	ComponentName      = "timer"
)

// The configuration parameters of the Timer component
type Config struct {
	TimeStep       string        `mapstructure:"time-step"`
	EnqueueItems   int           `mapstructure:"enqueue-items"`
	EnqueueTimeout time.Duration `mapstructure:"enqueue-timeout"`
	TracerUrl      string        `mapstructure:"tracer-url"`
	TracerTenant   string        `mapstructure:"tracer-tenant"`
}

// Add application-specific config parameters to flagset
func (cfg *Config) GetConfigFlagSet(fs *pflag.FlagSet) {
	fs.String(TIME_STEP_ARG_NAME, TIME_STEP_DEFAULT, TIME_STEP_HELP)
}

func (cfg *Config) LoadConfig(fs *pflag.FlagSet) error {
	if err := apprun.LoadConfigWithDefaultViper(fs, cfg, buildinfo.BuildInfo.AppName()+"_timer"); err != nil {
		return fmt.Errorf("failed to load config. %w", err)
	}
	return nil
}

var _ apprun.Configurer = (*Config)(nil)
