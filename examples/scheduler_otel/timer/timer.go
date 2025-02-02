package timer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/tombenke/go-12f-common/healthcheck"
	logger "github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/middleware"
	mw_inner "github.com/tombenke/go-12f-common/middleware/inner"

	"github.com/tombenke/go-12f-common/must"
	"github.com/tombenke/go-12f-common/tracing"

	"github.com/tombenke/go-12f-common/examples/scheduler_otel/buildinfo"
)

type TimerRequest struct {
	Ctx         context.Context
	CurrentTime time.Time
	JobID       string
}

// The Timer component
type Timer struct {
	config  *Config
	appWg   *sync.WaitGroup
	err     error
	rootCtx context.Context
	sendCh  chan TimerRequest

	ticker  *time.Ticker
	tr      trace.Tracer
	trDefer func()
}

type TimerStat struct {
	Duration time.Duration
	Success  int
	Failed   []error
}

var (
	ErrEnqueueRequests = errors.New("enqueue requests failed")
)

// Create a new Timer instance
func NewTimer(rootCtx context.Context, config *Config, currentTimeCh chan TimerRequest) (*Timer, error) {
	rootCtx, _ = logger.FromContext(rootCtx, logrus.Fields{logger.KeyService: ComponentName})
	return &Timer{
		config:  config,
		err:     healthcheck.ServiceNotAvailableError{},
		rootCtx: rootCtx,
		sendCh:  currentTimeCh,
	}, nil
}

// Startup the Timer component
func (t *Timer) Startup(wg *sync.WaitGroup) error {
	_, log := logger.FromContext(t.rootCtx, logrus.Fields{})
	t.appWg = wg
	wg.Add(1)
	log.Debugf("Timer: Startup")
	log.Debugf("Timer: config: %+v", t.config)

	tickerDuration := must.MustVal(time.ParseDuration(t.config.TimeStep))
	t.ticker = time.NewTicker(tickerDuration)
	var err error
	t.tr, t.trDefer, err = t.InitTracer(context.Background())
	if err != nil {
		return err
	}
	go t.Run()
	return nil
}

// Shutdown the Timer Component
func (t *Timer) Shutdown() error {
	_, log := logger.FromContext(t.rootCtx, logrus.Fields{})
	log.Debugf("Timer: Shutdown")

	// The component is not ready any more
	t.err = healthcheck.ServiceNotAvailableError{}

	// rootCtx already (must be) cancelled by the Application

	t.trDefer()

	return nil
}

// Run the component's processing logic within this function as a go-routine
func (t *Timer) Run() {
	_, log := logger.FromContext(t.rootCtx, logrus.Fields{})
	defer t.appWg.Done()
	defer log.Debugf("Timer: Stopped")

	// The component is working properly
	t.err = nil
	stepCounter := 0

	for {
		stepCounter++
		select {
		// Distribute the current simulation time value
		case currentTime := <-t.ticker.C:
			log.Debugf("Timer: tick: %v", currentTime)
			stat, err := t.runEnqueueRequests(t.rootCtx, currentTime, stepCounter)
			logLevel := logrus.InfoLevel
			if err != nil {
				logLevel = logrus.ErrorLevel
			}
			log.WithFields(logrus.Fields{
				"duration": stat.Duration, "success": stat.Success, "failed": len(stat.Failed), logger.KeyError: err,
			}).Log(logLevel, "EnqueueRequests")

		// Catch the shutdown signal
		case <-t.rootCtx.Done():
			log.Debugf("Timer: Shutting down")
			return
		}
	}
}

func (t *Timer) enqueueRequests(ctx context.Context, currentTime time.Time, stepCounter int) (TimerStat, error) {
	stat := TimerStat{}
	for j := 0; j < t.config.EnqueueItems; j++ {
		jobID := fmt.Sprintf("job#%d/%d", stepCounter, j)
		err := t.enqueueRequest(ctx, currentTime, jobID)
		if err != nil {
			stat.Failed = append(stat.Failed, err)
		} else {
			stat.Success++
		}
	}
	var err error
	if len(stat.Failed) > 0 {
		// TODO implement error list handling with ';' separator
		err = logger.Wrap(ErrEnqueueRequests, errors.Join(stat.Failed...))
	}

	return stat, err
}

var (
	ErrEnqueueTimeout  = errors.New("enqueue timeout")
	ErrEnqueueShutdown = errors.New("enqueue shutdown")
)

func (t *Timer) enqueueRequest(ctx context.Context, currentTime time.Time, jobID string) error {
	ctx, _ = logger.FromContext(ctx, logrus.Fields{"job_id": jobID})
	timeout := time.After(t.config.EnqueueTimeout)

	// Outgoing request trough channel: start a new Span
	// Skipping making new log, metric and trace (but possible)
	spanParent := trace.SpanFromContext(ctx).SpanContext()
	ctx, spanChild := t.tr.Start(ctx, "EnqueueRequest "+jobID,
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer spanChild.End()
	// Log Span parent, in order to see the hierarchy in the logs
	ctx, log := logger.FromContext(ctx, logrus.Fields{"spanParentID": spanParent.SpanID(), "spanID": spanChild.SpanContext().SpanID()})
	log.Debug("EnqueueRequest")

	select {
	case t.sendCh <- TimerRequest{
		Ctx:         ctx,
		CurrentTime: currentTime,
		JobID:       jobID,
	}: // The request is enqueued
	case <-timeout: // The request is dropped
		err := ErrEnqueueTimeout
		spanChild.RecordError(err, trace.WithStackTrace(true))
		spanChild.SetStatus(codes.Error, err.Error())
		return err
	case <-ctx.Done():
		err := ErrEnqueueShutdown
		spanChild.RecordError(err, trace.WithStackTrace(true))
		spanChild.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (t *Timer) runEnqueueRequests(ctx context.Context, currentTime time.Time, stepCounter int) (TimerStat, error) {
	_, log := logger.FromContext(ctx, logrus.Fields{})
	meter := middleware.GetMeter(buildinfo.BuildInfo, log)
	jobType := "backend"
	jobName := "enqueue_requests"
	appName := buildinfo.BuildInfo.AppName()
	componentName := ComponentName
	jobID := fmt.Sprintf("%s#%d", componentName, stepCounter)

	timerStat, err := mw_inner.InternalMiddlewareChain(
		mw_inner.TryCatch[TimerStat](),
		mw_inner.Span[TimerStat](t.tr, jobID),
		mw_inner.Logger[TimerStat](logrus.Fields{
			// Already set in the root context
			// logger.KeyApp:     appName,
			// logger.KeyService: componentName,
			"job_type": jobType,
			"job_name": jobName,
			"job_id":   jobID,
		}, logrus.InfoLevel, logrus.DebugLevel),
		mw_inner.Metrics[TimerStat](ctx, meter, "enqueue_requests", "Enqueue Requests", map[string]string{
			logger.KeyApp:     appName,
			logger.KeyService: componentName,
			"job_type":        jobType,
			"job_name":        jobName,
		}, middleware.FirstErr),
		mw_inner.TryCatch[TimerStat](),
	)(func(ctx context.Context) (TimerStat, error) {
		return t.enqueueRequests(ctx, currentTime, stepCounter)
	})(ctx)

	reportLevel := logrus.InfoLevel
	if err != nil {
		reportLevel = logrus.ErrorLevel
	}
	log.WithFields(logrus.Fields{"duration": timerStat.Duration, logger.KeyError: err}).Log(reportLevel, "RunEnqueueRequests")

	return timerStat, err
}

// Check if the component is ready to provide its services
func (t *Timer) Check() error {
	_, log := logger.FromContext(t.rootCtx, logrus.Fields{})
	log.Infof("Timer: Check")
	return t.err
}

func (t *Timer) InitTracer(ctx context.Context) (trace.Tracer, func(), error) {
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
