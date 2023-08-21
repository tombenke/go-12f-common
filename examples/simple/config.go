package main

import (
	"flag"
	"github.com/tombenke/go-12f-common/env"
)

const (
	StartupDelayHelp    = "The delay of startup process in seconds"
	StartupDelayEnvVar  = "STARTUP_DELAY"
	StartupDelayDefault = "30"
)

// The configuration parameters of the Application
type Config struct {
	StartupDelay int
}

// Add application-specific config parameters to flagset
func (cfg *Config) GetConfigFlagSet(fs *flag.FlagSet) {
	fs.IntVar(&cfg.StartupDelay, "startup-delay", int(env.GetEnvWithDefaultUint(StartupDelayEnvVar, StartupDelayDefault)), StartupDelayHelp)
}
