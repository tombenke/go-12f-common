package main

import (
	"context"
	"sync"
	"time"

	"github.com/sagikazarmark/slog-shim"
	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
)

type Application struct {
	config *Config
	err    error
}

func (a *Application) Components(ctx context.Context) []apprun.ComponentLifecycleManager {
	return nil
}

func (a *Application) AfterStartup(ctx context.Context, wg *sync.WaitGroup) error {
	_, logger := a.getLogger(ctx)
	logger.Info("AfterStartup")

	// After 3 seconds set the application readiness status to O.K.
	<-time.After(a.config.StartupDelay)
	logger.Info("Application should be healthy now", "after", a.config.StartupDelay)
	a.err = nil
	return nil
}

func (a *Application) BeforeShutdown(ctx context.Context) error {
	_, logger := a.getLogger(ctx)
	logger.Info("BeforeShutdown")
	a.err = healthcheck.ServiceNotAvailableError{}
	return nil
}

func (a *Application) Check(ctx context.Context) error {
	_, logger := a.getLogger(ctx)
	logger.Info("Check")
	return a.err
}

func (a *Application) getLogger(ctx context.Context) (context.Context, *slog.Logger) {
	return log.With(ctx, "app", "SimpleApplication")
}

func NewApplication(config *Config) (apprun.Application, error) {
	return &Application{
		err:    healthcheck.ServiceNotAvailableError{},
		config: config,
	}, nil
}
