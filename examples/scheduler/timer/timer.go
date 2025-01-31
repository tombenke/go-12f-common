package timer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tombenke/go-12f-common/healthcheck"
	internal_slog "github.com/tombenke/go-12f-common/slog"
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
	logger := internal_slog.GetFromContextOrDefault(ctx).With("component", "Timer")
	logger.Debug("Startup", "config", t.config)

	tickerDuration, err := time.ParseDuration(t.config.TimeStep)
	if err != nil {
		return fmt.Errorf("failed to parse TimeStep duration from config. %w", err)
	}

	t.ticker = time.NewTicker(tickerDuration)
	go t.Run(ctx)
	return nil
}

// Shutdown the Timer Component
func (t *Timer) Shutdown(ctx context.Context) error {
	internal_slog.DebugContext(ctx, "Shutdown", "component", "Timer")

	// The component is not ready any more
	t.err = healthcheck.ServiceNotAvailableError{}

	close(t.doneCh)
	return nil
}

// Run the component's processing logic within this function as a go-routine
func (t *Timer) Run(ctx context.Context) {
	logger := internal_slog.GetFromContextOrDefault(ctx).With("component", "Timer")
	defer t.appWg.Done()
	defer logger.DebugContext(ctx, "Stopped")

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
	internal_slog.InfoContext(ctx, "Check", "component", "Timer")
	return t.err
}
