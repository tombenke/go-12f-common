package worker

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/tombenke/go-12f-common/healthcheck"
	logger "github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/middleware"
	mw_inner "github.com/tombenke/go-12f-common/middleware/inner"
	"github.com/tombenke/go-12f-common/tracing"

	"github.com/tombenke/go-12f-common/examples/scheduler_otel/buildinfo"
	"github.com/tombenke/go-12f-common/examples/scheduler_otel/timer"
)

// The Worker component
type Worker struct {
	config    *Config
	appWg     *sync.WaitGroup
	err       error
	rootCtx   context.Context
	receiveCh chan timer.TimerRequest

	randNum rand.Rand
	tr      trace.Tracer
	trDefer func()
}

type WorkerStat struct {
	Duration time.Duration
}

var (
	ErrReceiveRequests = errors.New("receive requests failed")
)

// Create a new Worker instance
func NewWorker(rootCtx context.Context, config *Config, currentTimeCh chan timer.TimerRequest) (*Worker, error) {
	rootCtx, _ = logger.FromContext(rootCtx, logrus.Fields{logger.KeyService: ComponentName})
	return &Worker{
		config:    config,
		err:       healthcheck.ServiceNotAvailableError{},
		rootCtx:   rootCtx,
		receiveCh: currentTimeCh,
		randNum:   *rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Initialize the config parameters, then evaluate the environment variables and bind them for CLI argument evaluation
func (t *Worker) GetConfigFlagSet(fs *pflag.FlagSet) {
	// Delegate the config parameter initialization and binding to the Config object
	t.config.GetConfigFlagSet(fs)
}

// Startup the Worker component
func (t *Worker) Startup(wg *sync.WaitGroup) error {
	_, log := logger.FromContext(t.rootCtx, logrus.Fields{})
	t.appWg = wg
	wg.Add(1)
	log.Debugf("Worker: Startup")
	log.Debugf("Worker: config: %+v", t.config)

	var err error
	t.tr, t.trDefer, err = t.InitTracer(context.Background(), t.config.TracerUrl)
	if err != nil {
		return err
	}
	go t.Run()
	return nil
}

// Shutdown the Worker Component
func (t *Worker) Shutdown() error {
	_, log := logger.FromContext(t.rootCtx, logrus.Fields{})
	log.Debugf("Worker: Shutdown")

	// The components is ready any more
	t.err = healthcheck.ServiceNotAvailableError{}

	// rootCtx already (must be) cancelled by the Application

	t.trDefer()

	return nil
}

// Run the component's processing logic within this function as a go-routine
func (t *Worker) Run() {
	_, log := logger.FromContext(t.rootCtx, logrus.Fields{})
	defer t.appWg.Done()
	defer log.Debugf("Worker: Stopped")
	jobNum := 0

	// The component is working properly
	t.err = nil

	for {
		jobNum++
		select {
		case timerRequest, ok := <-t.receiveCh:
			if !ok {
				err := ErrReceiveRequests
				log.Error("Worker: ", logger.KeyError, err)
			} else {
				stat, err := t.runBusinessLogic(timerRequest.Ctx, timerRequest.JobID)
				logLevel := logrus.InfoLevel
				if err != nil {
					logLevel = logrus.ErrorLevel
				}
				log.WithFields(logrus.Fields{
					"duration": stat.Duration, logger.KeyError: err,
				}).Log(logLevel, "ProcessRequest")
			}
		case <-t.rootCtx.Done():
			log.Debugf("Worker: Shutting down")
			return
		}
	}
}

func (t *Worker) businessLogic(ctx context.Context, jobID string) (WorkerStat, error) {
	_, log := logger.FromContext(ctx, logrus.Fields{})
	duration := t.config.ProcessMinDuration + time.Duration(
		t.randNum.Int63n(int64(t.config.ProcessMaxDuration-t.config.ProcessMinDuration)),
	)
	log.Infof("Worker: Processing job: %s, duration: %v", jobID, duration)
	time.Sleep(duration)

	return WorkerStat{duration}, nil
}

func (t *Worker) runBusinessLogic(ctx context.Context, jobID string) (WorkerStat, error) {
	_, log := logger.FromContext(ctx, logrus.Fields{})
	meter := middleware.GetMeter(buildinfo.BuildInfo, log)
	jobType := "backend"
	jobName := "process_requests"
	appName := buildinfo.BuildInfo.AppName()
	componentName := ComponentName

	workerStat, err := mw_inner.InternalMiddlewareChain(
		mw_inner.TryCatch[WorkerStat](),
		mw_inner.Span[WorkerStat](t.tr, jobID),
		mw_inner.Logger[WorkerStat](logrus.Fields{
			// Already set in the root context
			// logger.KeyApp:     appName,
			// logger.KeyService: componentName,
			"job_type": jobType,
			"job_name": jobName,
			"job_id":   jobID,
		}, logrus.InfoLevel, logrus.DebugLevel),
		mw_inner.Metrics[WorkerStat](ctx, meter, "process_request", "Process Request", map[string]string{
			logger.KeyApp:     appName,
			logger.KeyService: componentName,
			"job_type":        jobType,
			"job_name":        jobName,
		}, middleware.FirstErr),
		mw_inner.TryCatch[WorkerStat](),
	)(func(ctx context.Context) (WorkerStat, error) {
		return t.businessLogic(ctx, jobID)
	})(ctx)

	reportLevel := logrus.InfoLevel
	if err != nil {
		reportLevel = logrus.ErrorLevel
	}
	log.WithFields(logrus.Fields{"duration": workerStat.Duration, logger.KeyError: err}).Log(reportLevel, "BUSINESS_LOGIC")

	return workerStat, err
}

// Check if the component is ready to provide its services
func (t *Worker) Check() error {
	_, log := logger.FromContext(t.rootCtx, logrus.Fields{})
	log.Infof("Worker: Check")
	return t.err
}

func (t *Worker) InitTracer(ctx context.Context, tracerUrl string) (trace.Tracer, func(), error) {
	_, log := logger.FromContext(ctx, logrus.Fields{})
	hostname, _ := os.Hostname() //nolint:errcheck // not important
	componentName := ComponentName
	deferFn := func() {}

	traceOptions := []otlptracehttp.Option{}
	if t.config.TracerUrl != "" {
		traceOptions = append(traceOptions, otlptracehttp.WithEndpointURL(t.config.TracerUrl))
		traceOptions = append(traceOptions, otlptracehttp.WithCompression(otlptracehttp.NoCompression))
		traceOptions = append(traceOptions, otlptracehttp.WithInsecure())
	}
	if t.config.TracerTenant != "" {
		traceOptions = append(traceOptions, otlptracehttp.WithHeaders(map[string]string{"X-Scope-OrgID": t.config.TracerTenant}))
	}
	traceExporter, err := tracing.OtlpProvider(ctx, traceOptions...)
	if err != nil {
		return nil, deferFn, logger.Wrap(errors.New("unable to get OtlpProvider"), err)
	}
	tp := tracing.InitTracer(traceExporter, sdktrace.AlwaysSample(),
		buildinfo.BuildInfo, componentName, hostname, "", log,
	)
	deferFn = func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error("Error shutting down tracer provider", logger.KeyService, componentName, logger.KeyError, err)
		}
	}

	tr := GetTracer(tp)

	return tr, deferFn, nil
}

func GetTracer(tp trace.TracerProvider) trace.Tracer {
	return tp.Tracer(
		buildinfo.BuildInfo.ModulePath(),
		trace.WithInstrumentationVersion(tracing.SemVersion()),
	)
}
