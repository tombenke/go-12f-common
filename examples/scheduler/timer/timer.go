package timer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tombenke/go-12f-common/v2/healthcheck"
	"github.com/tombenke/go-12f-common/v2/log"
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
func NewTimer(config *Config, currentTimeCh chan time.Time) *Timer {
	doneCh := make(chan interface{})
	return &Timer{config: config, err: healthcheck.ServiceNotAvailableError{}, doneCh: doneCh, currentTimeCh: currentTimeCh}
}

// Startup the Timer component
func (t *Timer) Startup(ctx context.Context, wg *sync.WaitGroup) error {
	t.appWg = wg
	wg.Add(1)
	ctx, logger := t.getLogger(ctx)
	logger.Debug("Startup", "config", t.config)

	tickerDuration, err := time.ParseDuration(t.config.TimeStep)
	if err != nil {
		return fmt.Errorf("failed to parse TimeStep duration from config. %w", err)
	}

	logger.Debug("Starting ticker", "duration", tickerDuration.String())
	t.ticker = time.NewTicker(tickerDuration)
	go t.run(ctx)
	return nil
}

// Shutdown the Timer Component
func (t *Timer) Shutdown(ctx context.Context) error {
	_, logger := t.getLogger(ctx)
	logger.Debug("Shutdown")

	// The component is not ready any more
	t.err = healthcheck.ServiceNotAvailableError{}

	close(t.doneCh)
	return nil
}

// run the component's processing logic within this function as a go-routine
func (t *Timer) run(ctx context.Context) {
	logger := log.GetFromContextOrDefault(ctx)
	defer t.appWg.Done()
	defer logger.Debug("Stopped")

	// The component is working properly
	t.err = nil

	for {
		select {
		// Distribute the current simulation time value
		case currentTime := <-t.ticker.C:
			logger.Debug("Tick", "currentTime", currentTime)
			t.currentTimeCh <- currentTime
			continue

		// Catch the shutdown signal
		case <-t.doneCh:
			logger.Debug("Shutting down")
			return
		}
	}
}

// Check if the component is ready to provide its services
func (t *Timer) Check(ctx context.Context) error {
	_, logger := t.getLogger(ctx)
	logger.Info("Check")
	return t.err
}

func (t *Timer) getLogger(ctx context.Context) (context.Context, *slog.Logger) {
	return log.With(ctx, "component", "Timer")
}
