package scheduler

import (
	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"sync"
	"time"
)

// The Scheduler component
type Scheduler struct {
	config        *Config
	appWg         *sync.WaitGroup
	err           error
	doneCh        chan interface{}
	currentTimeCh chan time.Time
}

// Create a new Scheduler instance
func NewScheduler(config *Config, currentTimeCh chan time.Time) (*Scheduler, error) {
	doneCh := make(chan interface{})
	return &Scheduler{config: config, err: healthcheck.ServiceNotAvailableError{}, doneCh: doneCh, currentTimeCh: currentTimeCh}, nil
}

// Initialize the config parameters, then evaluate the environment variables and bind them for CLI argument evaluation
func (t *Scheduler) GetConfigFlagSet(fs *pflag.FlagSet) {
	// Delegate the config parameter initialization and binding to the Config object
	t.config.GetConfigFlagSet(fs)
}

// Startup the Scheduler component
func (t *Scheduler) Startup(wg *sync.WaitGroup) error {
	t.appWg = wg
	wg.Add(1)
	log.Logger.Debugf("Scheduler: Startup")
	log.Logger.Debugf("Scheduler: config: %+v", t.config)
	go t.Run()
	return nil
}

// Shutdown the Scheduler Component
func (t *Scheduler) Shutdown() error {
	log.Logger.Debugf("Scheduler: Shutdown")

	// The components is ready any more
	t.err = healthcheck.ServiceNotAvailableError{}

	close(t.doneCh)
	return nil
}

// Run the component's processing logic within this function as a go-routine
func (t *Scheduler) Run() {
	defer t.appWg.Done()
	defer log.Logger.Debugf("Scheduler: Stopped")

	// The component is working properly
	t.err = nil

	for {
		select {
		case currentTime := <-t.currentTimeCh:
			// TODO: Implement the processing feature
			log.Logger.Debugf("Scheduler: currentTime: %v", currentTime)
			continue

		case <-t.doneCh:
			log.Logger.Debugf("Scheduler: Shutting down")
			return
		}
	}
}

// Check if the component is ready to provide its services
func (t *Scheduler) Check() error {
	log.Logger.Infof("Scheduler: Check")
	return t.err
}
