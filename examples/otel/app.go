package main

import (
	"context"
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

func (a *Application) Startup(ctx context.Context, wg *sync.WaitGroup) error {
	a.wg = wg
	log.Logger.Infof("Application Startup")
	a.err = nil
	return nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	log.Logger.Infof("Application Shutdown")
	a.err = healthcheck.ServiceNotAvailableError{}
	return nil
}

func (a *Application) Check(ctx context.Context) error {
	log.Logger.Infof("Application Check")
	return a.err
}

func NewApplication(config *Config) (apprun.LifecycleManager, error) {
	return &Application{err: healthcheck.ServiceNotAvailableError{}, config: config}, nil
}
