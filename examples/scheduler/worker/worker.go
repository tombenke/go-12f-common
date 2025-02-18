package worker

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/v2/healthcheck"
	"github.com/tombenke/go-12f-common/v2/log"
	metric_api "go.opentelemetry.io/otel/metric"

	"github.com/tombenke/go-12f-common/v2/examples/scheduler/model"
	"github.com/tombenke/go-12f-common/v2/oti"
)

const (
	RunCountName = "run"
)

var (
	processMinDuration = 100 * time.Millisecond
	processMaxDuration = 500 * time.Millisecond
)

// The Worker component
type Worker struct {
	config        *Config
	appWg         *sync.WaitGroup
	err           error
	doneCh        chan interface{}
	currentTimeCh chan model.TimerRequest
	runCount      metric_api.Int64Counter
}

// Create a new Worker instance
func NewWorker(config *Config, currentTimeCh chan model.TimerRequest) *Worker {
	doneCh := make(chan interface{})
	return &Worker{
		config:        config,
		err:           healthcheck.ServiceNotAvailableError{},
		doneCh:        doneCh,
		currentTimeCh: currentTimeCh,
	}
}

// Initialize the config parameters, then evaluate the environment variables and bind them for CLI argument evaluation
func (t *Worker) GetConfigFlagSet(fs *pflag.FlagSet) {
	// Delegate the config parameter initialization and binding to the Config object
	t.config.GetConfigFlagSet(fs)
}

// Startup the Worker component
func (t *Worker) Startup(ctx context.Context, wg *sync.WaitGroup) error {
	_, logger := t.getLogger(ctx)

	t.appWg = wg
	wg.Add(1)
	logger.Debug("Startup", "config", t.config)

	var err error
	t.runCount, err = oti.Int64CounterGetInstrument(
		RunCountName,
		metric_api.WithDescription("The number of times the worker run"),
	)
	if err != nil {
		logger.Error("failed runCount meter creation",
			oti.KeyError, err, oti.KeyMetricName, RunCountName)
		panic(err)
	}

	// Run the worker
	go t.run(ctx)

	return nil
}

// Shutdown the Worker Component
func (t *Worker) Shutdown(ctx context.Context) error {
	_, logger := t.getLogger(ctx)
	logger.Debug("Shutdown")

	// The components is ready any more
	t.err = healthcheck.ServiceNotAvailableError{}

	close(t.doneCh)
	return nil
}

// run the component's processing logic within this function as a go-routine
func (t *Worker) run(ctx context.Context) {
	_, logger := t.getLogger(ctx)
	defer t.appWg.Done()
	defer logger.Debug("Stopped")

	// The component is working properly
	t.err = nil

	for {
		select {
		case currentTime := <-t.currentTimeCh:
			logger.Debug("Tick", "current.time", currentTime.CurrentTime)
			t.runCount.Add(ctx, 1)
			obsProcessTimerRequest(
				t, t.processTimerRequest,
			)(ctx, currentTime)
			// TODO remove continue, because it is not necessary (not C)
			continue
		case <-t.doneCh:
			logger.Debug("Shutting down")
			return
		}
	}
}

func (t *Worker) processTimerRequest(ctx context.Context, timerRequest model.TimerRequest) (time.Duration, error) {
	_, logger := t.getLogger(ctx)
	duration := processMinDuration + time.Duration(
		rand.Int64N(int64(processMaxDuration-processMinDuration)),
	)
	time.Sleep(duration)
	logger.Info("Processing", "current.time", timerRequest.CurrentTime, "processing.duration", duration)

	return duration, nil
}

// Check if the component is ready to provide its services
func (t *Worker) Check(ctx context.Context) error {
	_, logger := t.getLogger(ctx)
	logger.Info("Check")
	return t.err
}

/*
getLogger returns a logger instance with the component name set

TODO delete this function, because:
* violates request-scoped logger and tracer principle
* multiple calls adds the component field multiple times
*/
func (t *Worker) getLogger(ctx context.Context) (context.Context, *slog.Logger) {
	return log.With(ctx, "component", t.ComponentName())
}

func (t *Worker) ComponentName() string {
	return "Worker"
}
