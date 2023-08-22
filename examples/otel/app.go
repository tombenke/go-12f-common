package main

import (
	"flag"
	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"sync"
)

type Application struct {
	config Config
	wg     *sync.WaitGroup
	err    error
}

func (a *Application) GetConfigFlagSet(fs *flag.FlagSet) {
	a.config.GetConfigFlagSet(fs)
}

func (a *Application) Startup(wg *sync.WaitGroup) {
	a.wg = wg
	log.Logger.Infof("Application Startup")
	a.err = nil
}

func (a *Application) Shutdown() {
	log.Logger.Infof("Application Shutdown")
	a.err = healthcheck.ServiceNotAvailableError{}
}

func (a *Application) Check() error {
	log.Logger.Infof("Application Check")
	return a.err
}

func NewApplication() (apprun.LifecycleManager, error) {
	return &Application{err: healthcheck.ServiceNotAvailableError{}, config: Config{}}, nil
}
