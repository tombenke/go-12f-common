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
	wg     *sync.WaitGroup
	err    error
}

func (a *Application) Startup(ctx context.Context, wg *sync.WaitGroup) error {
	_, logger := a.getLogger(ctx)
	a.wg = wg
	logger.Info("Startup")

	// After 3 seconds set the application readiness status to O.K.
	<-time.After(time.Duration(a.config.StartupDelay) * time.Second)
	a.err = nil
	return nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	_, logger := a.getLogger(ctx)
	logger.Info("Shutdown")
	a.err = healthcheck.ServiceNotAvailableError{}
	return nil
}

func (a *Application) Check(ctx context.Context) error {
	_, logger := a.getLogger(ctx)
	logger.Info("Check")
	return a.err
}

func (a *Application) getLogger(ctx context.Context) (context.Context, *slog.Logger) {
	return log.With(ctx, "app", "Application")
}

func NewApplication(config *Config) (apprun.LifecycleManager, error) {
	return &Application{
		err:    healthcheck.ServiceNotAvailableError{},
		config: config,
	}, nil
}
