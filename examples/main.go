package main

import (
	"flag"
	"github.com/tombenke/go-12f-common/app"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"sync"
	"time"

	"github.com/tombenke/go-12f-common/env"
)

const (
	ServerPortHelp    = "The HTTP port of the server"
	ServerPortEnvVar  = "SERVER_PORT"
	ServerPortDefault = "8081"
)

// The configuration parameters of the ExampleApp
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

func (a *ExampleApp) Startup(wg *sync.WaitGroup) {
	a.wg = wg
	log.Logger.Infof("ExampleApp Startup")

	// After 3 seconds set the application readiness status to O.K.
	<-time.After(30 * time.Second)
	a.err = nil
}

func (a *ExampleApp) Shutdown() {
	log.Logger.Infof("ExampleApp Shutdown")
}

func (a *ExampleApp) Check() error {
	log.Logger.Infof("ExampleApp Check")
	return a.err
}

func NewExampleApp() (app.LifecycleManager, error) {
	return &ExampleApp{err: healthcheck.ServiceNotAvailableError{}, config: Config{}}, nil
}

func main() {

	// Make and run an application via ApplicationRunner
	app.MakeAndRun(NewExampleApp)
}
