package worker

import (
	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"sync"
	"time"
)

// The Worker component
type Worker struct {
	config        *Config
	appWg         *sync.WaitGroup
	err           error
	doneCh        chan interface{}
	currentTimeCh chan time.Time
}

// Create a new Worker instance
func NewWorker(config *Config, currentTimeCh chan time.Time) (*Worker, error) {
	doneCh := make(chan interface{})
	return &Worker{config: config, err: healthcheck.ServiceNotAvailableError{}, doneCh: doneCh, currentTimeCh: currentTimeCh}, nil
}

// Initialize the config parameters, then evaluate the environment variables and bind them for CLI argument evaluation
func (t *Worker) GetConfigFlagSet(fs *pflag.FlagSet) {
	// Delegate the config parameter initialization and binding to the Config object
	t.config.GetConfigFlagSet(fs)
}

// Startup the Worker component
func (t *Worker) Startup(wg *sync.WaitGroup) error {
	t.appWg = wg
	wg.Add(1)
	log.Logger.Debugf("Worker: Startup")
	log.Logger.Debugf("Worker: config: %+v", t.config)
	go t.Run()
	return nil
}

// Shutdown the Worker Component
func (t *Worker) Shutdown() error {
	log.Logger.Debugf("Worker: Shutdown")

	// The components is ready any more
	t.err = healthcheck.ServiceNotAvailableError{}

	close(t.doneCh)
	return nil
}

// Run the component's processing logic within this function as a go-routine
func (t *Worker) Run() {
	defer t.appWg.Done()
	defer log.Logger.Debugf("Worker: Stopped")

	// The component is working properly
	t.err = nil

	for {
		select {
		case currentTime := <-t.currentTimeCh:
			// TODO: Implement the processing feature
			log.Logger.Debugf("Worker: currentTime: %v", currentTime)
			continue

		case <-t.doneCh:
			log.Logger.Debugf("Worker: Shutting down")
			return
		}
	}
}

// Check if the component is ready to provide its services
func (t *Worker) Check() error {
	log.Logger.Infof("Worker: Check")
	return t.err
}
