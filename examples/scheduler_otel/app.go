package main

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/healthcheck"
	logger "github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"

	"github.com/tombenke/go-12f-common/examples/scheduler_otel/timer"
	"github.com/tombenke/go-12f-common/examples/scheduler_otel/worker"
)

type Application struct {
	config        *Config
	wg            *sync.WaitGroup
	err           error
	currentTimeCh chan timer.TimerRequest
	rootCtx       context.Context
	rootCtxCancel context.CancelCauseFunc

	// The internal components of the application
	components []apprun.LifecycleManager
}

var (
	ErrShutdownReceived = errors.New("shutdown received")
)

func (a *Application) Startup(wg *sync.WaitGroup) error {
	_, log := logger.FromContext(a.rootCtx)
	log.Info("Application Startup")

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
	_, log := logger.FromContext(a.rootCtx)
	log.Info("Application Shutdown")
	a.rootCtxCancel(ErrShutdownReceived)

	a.err = healthcheck.ServiceNotAvailableError{}

	// Shutdown the internal components
	for c := range a.components {
		a.components[c].Shutdown()
	}

	// Close channels is the last step, in order to avoid
	// panic on sending messages to already closed channels
	a.closeChannels()

	return nil
}

// Close the inter-component communication channels
func (a *Application) closeChannels() {
	if a.currentTimeCh != nil {
		close(a.currentTimeCh)
	}
}

func (a *Application) Check() error {
	_, log := logger.FromContext(a.rootCtx)
	log.Info("Application Check")
	return a.err
}

func NewApplication(config *Config) (apprun.LifecycleManager, error) {
	appName := "scheduler_otel"
	log := logger.GetLogger(appName, slog.LevelDebug).With(logger.KeyApp, appName)

	rootCtx, rootCtxCancel := context.WithCancelCause(context.Background())
	rootCtx = logger.NewContext(rootCtx, log)
	log.Info("Application.Config", "config", *config)
	// Create channel(s) for inter-component communication
	currentTimeCh := make(chan timer.TimerRequest, config.QueueSize)

	return &Application{
		config:        config,
		err:           healthcheck.ServiceNotAvailableError{},
		rootCtx:       rootCtx,
		rootCtxCancel: rootCtxCancel,
		components: []apprun.LifecycleManager{
			must.MustVal(timer.NewTimer(rootCtx, &config.timer, currentTimeCh)),
			must.MustVal(worker.NewWorker(rootCtx, &config.worker, currentTimeCh)),
		},
	}, nil
}

/*
	hostname, _ := os.Hostname() //nolint:errcheck // not important
	deferFn := func() {}

	traceOptions := []otlptracehttp.Option{}
	if serverConfig.GetTracerUrl() != "" {
		traceOptions = append(traceOptions, otlptracehttp.WithEndpointURL(serverConfig.GetTracerUrl()))
		traceOptions = append(traceOptions, otlptracehttp.WithCompression(otlptracehttp.NoCompression))
		traceOptions = append(traceOptions, otlptracehttp.WithInsecure())
	}
	traceExporter, err := tracing.OtlpProvider(ctx, traceOptions...)
	if err != nil {
		return deferFn, logger.Wrap(errors.New("unable to get OtlpProvider"), err)
	}
	tp := tracing.InitTracer(traceExporter, sdktrace.AlwaysSample(),
		buildinfo, service.Name(), hostname, "", log,
	)
	deferFn = func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error("Error shutting down tracer provider", logger.KeyService, service.Name(), logger.KeyError, err)
		}
	}
	tr := tp.Tracer(
		buildinfo.ModulePath(),
		trace.WithInstrumentationVersion(tracing.SemVersion()),
	)
*/
