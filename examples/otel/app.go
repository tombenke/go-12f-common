package main

import (
	"sync"

	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
)

type Application struct {
	config *Config
	wg     *sync.WaitGroup
	err    error
}

func (a *Application) Startup(wg *sync.WaitGroup) error {
	a.wg = wg
	log.Logger.Infof("Application Startup")
	a.err = nil
	return nil
}

func (a *Application) Shutdown() error {
	log.Logger.Infof("Application Shutdown")
	a.err = healthcheck.ServiceNotAvailableError{}
	return nil
}

func (a *Application) Check() error {
	log.Logger.Infof("Application Check")
	return a.err
}

func NewApplication(config *Config) (apprun.LifecycleManager, error) {
	return &Application{err: healthcheck.ServiceNotAvailableError{}, config: config}, nil
}
