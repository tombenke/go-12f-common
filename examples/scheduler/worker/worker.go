package worker

import (
	"context"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/healthcheck"
	internal_slog "github.com/tombenke/go-12f-common/slog"
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
func NewWorker(config *Config, currentTimeCh chan time.Time) *Worker {
	doneCh := make(chan interface{})
	return &Worker{config: config, err: healthcheck.ServiceNotAvailableError{}, doneCh: doneCh, currentTimeCh: currentTimeCh}
}

// Initialize the config parameters, then evaluate the environment variables and bind them for CLI argument evaluation
func (t *Worker) GetConfigFlagSet(fs *pflag.FlagSet) {
	// Delegate the config parameter initialization and binding to the Config object
	t.config.GetConfigFlagSet(fs)
}

// Startup the Worker component
func (t *Worker) Startup(ctx context.Context, wg *sync.WaitGroup) error {
	t.appWg = wg
	wg.Add(1)
	internal_slog.DebugContext(ctx, "Startup", "component", "Worker", "config", t.config)
	go t.Run(ctx)
	return nil
}

// Shutdown the Worker Component
func (t *Worker) Shutdown(ctx context.Context) error {
	internal_slog.DebugContext(ctx, "Shutdown", "component", "Worker")

	// The components is ready any more
	t.err = healthcheck.ServiceNotAvailableError{}

	close(t.doneCh)
	return nil
}

// Run the component's processing logic within this function as a go-routine
func (t *Worker) Run(ctx context.Context) {
	logger := internal_slog.GetFromContextOrDefault(ctx).With("component", "Worker")
	defer t.appWg.Done()
	defer logger.Debug("Stopped")

	// The component is working properly
	t.err = nil

	for {
		select {
		case currentTime := <-t.currentTimeCh:
			// TODO: Implement the processing feature
			logger.Debug("Tick", "currentTime", currentTime)
			continue

		case <-t.doneCh:
			logger.Debug("Shutting down")
			return
		}
	}
}

// Check if the component is ready to provide its services
func (t *Worker) Check(ctx context.Context) error {
	internal_slog.InfoContext(ctx, "Check")
	return t.err
}
