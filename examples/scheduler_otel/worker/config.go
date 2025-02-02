package worker

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/examples/scheduler_otel/buildinfo"
)

const (
	ComponentName = "worker"
)

// The configuration parameters of the Worker component
type Config struct {
	TracerUrl          string        `mapstructure:"tracer-url"`
	TracerTenant       string        `mapstructure:"tracer-tenant"`
	ProcessMinDuration time.Duration `mapstructure:"process-min-duration"`
	ProcessMaxDuration time.Duration `mapstructure:"process-max-duration"`
}

// Add application-specific config parameters to flagset
func (cfg *Config) GetConfigFlagSet(fs *pflag.FlagSet) {
}

func (cfg *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := apprun.LoadConfigWithDefaultViper(flagSet, cfg, buildinfo.BuildInfo.AppName()+"_worker"); err != nil {
		return fmt.Errorf("failed to load worker config. %w", err)
	}
	return nil
}
