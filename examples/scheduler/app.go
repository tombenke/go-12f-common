package main

import (
	"sync"
	"time"

	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/examples/scheduler/scheduler"
	"github.com/tombenke/go-12f-common/examples/scheduler/timer"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"
)

type Application struct {
	config        *Config
	wg            *sync.WaitGroup
	err           error
	currentTimeCh chan (time.Time)

	// The internal components of the application
	components []apprun.LifecycleManager
}

func (a *Application) Startup(wg *sync.WaitGroup) error {
	log.Logger.Infof("Application Startup")

	// Inject the central waitgroup into the application object
	a.wg = wg

	// Startup the internal components
	for c := range a.components {
		a.components[c].Startup(a.wg)
	}

	// Check if Application is ready to serve service calls
	_ = a.Check()

	a.err = nil
	return nil
}

func (a *Application) Shutdown() error {
	log.Logger.Infof("Application Shutdown")
	a.closeChannels()

	a.err = healthcheck.ServiceNotAvailableError{}

	// Shutdown the internal components
	for c := range a.components {
		a.components[c].Shutdown()
	}
	return nil
}

// Close the inter-component communication channels
func (a *Application) closeChannels() {
	if a.currentTimeCh != nil {
		close(a.currentTimeCh)
	}
}

func (a *Application) Check() error {
	log.Logger.Infof("Application Check")
	return a.err
}

func NewApplication(config *Config) (apprun.LifecycleManager, error) {
	log.Logger.Infof("Application.Config: %+v", *config)
	// Create channel(s) for inter-component communication
	currentTimeCh := make(chan (time.Time))

	return &Application{
		config: config,
		err:    healthcheck.ServiceNotAvailableError{},
		components: []apprun.LifecycleManager{
			must.MustVal(timer.NewTimer(&config.timer, currentTimeCh)),
			must.MustVal(scheduler.NewScheduler(&config.scheduler, currentTimeCh)),
		},
	}, nil
}
