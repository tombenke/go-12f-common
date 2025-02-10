package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/buildinfo"
	"github.com/tombenke/go-12f-common/examples/scheduler/timer"
	"github.com/tombenke/go-12f-common/examples/scheduler/worker"
	"github.com/tombenke/go-12f-common/log"
)

//var Version string = "0.0.0"

type Application struct {
	config        *Config
	currentTimeCh chan (time.Time)

	// The internal components of the application
	components []apprun.ComponentLifecycleManager
}

func (a *Application) Components(ctx context.Context) []apprun.ComponentLifecycleManager {
	return a.components
}

func (a *Application) AfterStartup(ctx context.Context, wg *sync.WaitGroup) error {
	a.getLogger(ctx).Info("AfterStartup")
	a.getLogger(ctx).Debug("BuildInfo", "AppName", buildinfo.AppName(), "Version", buildinfo.Version())
	return nil
}

func (a *Application) BeforeShutdown(ctx context.Context) error {
	a.getLogger(ctx).Info("BeforeShutdown")
	a.closeChannels()
	return nil
}

// Close the inter-component communication channels
func (a *Application) closeChannels() {
	if a.currentTimeCh != nil {
		close(a.currentTimeCh)
	}
}

func (a *Application) getLogger(ctx context.Context) *slog.Logger {
	return log.GetFromContextOrDefault(ctx).With("app", "Application")
}

func NewApplication(config *Config) (apprun.Application, error) {
	slog.Info("Creating Application", "config", *config)
	// Create channel(s) for inter-component communication
	currentTimeCh := make(chan (time.Time))

	return &Application{
		config:        config,
		currentTimeCh: currentTimeCh,
		components: []apprun.ComponentLifecycleManager{
			timer.NewTimer(&config.timer, currentTimeCh),
			worker.NewWorker(&config.worker, currentTimeCh),
		},
	}, nil
}
