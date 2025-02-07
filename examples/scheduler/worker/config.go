package worker

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/config"
)

const ()

// The configuration parameters of the Worker component
type Config struct {
}

// Add application-specific config parameters to flagset
func (cfg *Config) GetConfigFlagSet(fs *pflag.FlagSet) {
}

func (cfg *Config) LoadConfig(flagSet *pflag.FlagSet) error {
	if err := config.LoadConfigWithDefaultViper(flagSet, cfg); err != nil {
		return fmt.Errorf("failed to load worker config. %w", err)
	}
	return nil
}
