package main

import (
	"flag"
	"github.com/tombenke/go-12f-common/app"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"sync"

	"github.com/tombenke/go-12f-common/env"
)

const (
	ServerPortHelp    = "The HTTP port of the server"
	ServerPortEnvVar  = "SERVER_PORT"
	ServerPortDefault = "8081"
)

type Config struct {
	ServerPort int64
}

// Add application-specific config parameters to flagset
func (cfg *Config) GetConfigFlagSet(fs *flag.FlagSet) {
	fs.Int64Var(&cfg.ServerPort, "server-port", int64(env.GetEnvWithDefaultUint(ServerPortEnvVar, ServerPortDefault)), ServerPortHelp)
}

type ExampleApp struct {
	config Config
	wg     *sync.WaitGroup
	err    error
}

func (a *ExampleApp) GetConfigFlagSet(fs *flag.FlagSet) {
	a.config.GetConfigFlagSet(fs)
}

func (a *ExampleApp) Startup() {
	log.Logger.Infof("ExampleApp Startup")
	////log.Logger.Infof("ar.app.config: %v", a.config)
}

func (a *ExampleApp) Shutdown() {
	log.Logger.Infof("ExampleApp Shutdown")
}

func (a *ExampleApp) Check() error {
	log.Logger.Infof("ExampleApp Check")
	return a.err
}

func NewExampleApp(wg *sync.WaitGroup) (app.LifecycleManager, error) {
	return &ExampleApp{wg: wg, err: healthcheck.ServiceNotAvailableError{}, config: Config{}}, nil
}

func main() {

	// Create a waitgroup for the application's sub-processes
	appWg := sync.WaitGroup{}

	// Make and run an application via ApplicationRunner
	app.MakeAndRun(&appWg, NewExampleApp)

	// Wait until the application has shut down
	appWg.Wait()
}
