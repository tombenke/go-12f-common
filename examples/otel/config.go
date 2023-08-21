package main

import (
	"flag"
	"github.com/tombenke/go-12f-common/env"
)

const (
	ServerPortHelp    = "The HTTP port of the server"
	ServerPortEnvVar  = "SERVER_PORT"
	ServerPortDefault = "8081"
)

// The configuration parameters of the ExampleApp
type Config struct {
	ServerPort int
}

// Add application-specific config parameters to flagset
func (cfg *Config) GetConfigFlagSet(fs *flag.FlagSet) {
	fs.IntVar(&cfg.ServerPort, "server-port", int(env.GetEnvWithDefaultUint(ServerPortEnvVar, ServerPortDefault)), ServerPortHelp)
}
