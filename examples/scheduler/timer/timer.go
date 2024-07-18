package timer

import (
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"
	"sync"
	"time"
)

// The Timer component
type Timer struct {
	config        *Config
	appWg         *sync.WaitGroup
	err           error
	doneCh        chan interface{}
	currentTimeCh chan time.Time
	ticker        *time.Ticker
}

// Create a new Timer instance
func NewTimer(config *Config, currentTimeCh chan time.Time) (*Timer, error) {
	doneCh := make(chan interface{})
	return &Timer{config: config, err: healthcheck.ServiceNotAvailableError{}, doneCh: doneCh, currentTimeCh: currentTimeCh}, nil
}

// Startup the Timer component
func (t *Timer) Startup(wg *sync.WaitGroup) error {
	t.appWg = wg
	wg.Add(1)
	log.Logger.Debugf("Timer: Startup")
	log.Logger.Debugf("Timer: config: %+v", t.config)

	tickerDuration := must.MustVal(time.ParseDuration(t.config.TimeStep))
	t.ticker = time.NewTicker(tickerDuration)
	go t.Run()
	return nil
}

// Shutdown the Timer Component
func (t *Timer) Shutdown() error {
	log.Logger.Debugf("Timer: Shutdown")

	// The component is not ready any more
	t.err = healthcheck.ServiceNotAvailableError{}

	close(t.doneCh)
	return nil
}

// Run the component's processing logic within this function as a go-routine
func (t *Timer) Run() {
	defer t.appWg.Done()
	defer log.Logger.Debugf("Timer: Stopped")

	// The component is working properly
	t.err = nil

	for {
		select {
		// Distribute the current simulation time value
		case currentTime := <-t.ticker.C:
			log.Logger.Debugf("Timer: tick: %v", currentTime)
			t.currentTimeCh <- currentTime
			continue

		// Catch the shutdown signal
		case <-t.doneCh:
			log.Logger.Debugf("Timer: Shutting down")
			return
		}
	}
}

// Check if the component is ready to provide its services
func (t *Timer) Check() error {
	log.Logger.Infof("Timer: Check")
	return t.err
}
