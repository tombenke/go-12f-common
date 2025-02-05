package main

import (
	"context"
	"log/slog"
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
	a.getLogger(ctx).Info("Startup")
	a.err = nil
	return nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	a.getLogger(ctx).Info("Shutdown")
	a.err = healthcheck.ServiceNotAvailableError{}
	return nil
}

func (a *Application) Check(ctx context.Context) error {
	a.getLogger(ctx).Info("Check")
	return a.err
}

func (a *Application) getLogger(ctx context.Context) *slog.Logger {
	return log.GetFromContextOrDefault(ctx).With("app", "Application")
}

func NewApplication(config *Config) (apprun.LifecycleManager, error) {
	return &Application{err: healthcheck.ServiceNotAvailableError{}, config: config}, nil
}
