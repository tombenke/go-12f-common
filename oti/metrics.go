package oti

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	client_prometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/tombenke/go-12f-common/v2/must"

	////"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	metric_api "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/tombenke/go-12f-common/v2/buildinfo"
)

const (
	// MetrAttrErr is the metric attribute for error
	MetrAttrErr           = attribute.Key("error")
	MetrAttrMethod        = semconv.HTTPRequestMethodKey
	MetrAttrUrl           = semconv.URLFullKey
	MetrAttrStatus        = semconv.HTTPResponseStatusCodeKey
	MetrAttrPathPattern   = semconv.URLTemplateKey
	MetrAttrPath          = semconv.URLPathKey
	MetrAttrHost          = semconv.URLDomainKey
	MetrAttrService       = semconv.ServiceNameKey
	MetrAttrTargetService = semconv.PeerServiceKey

	MetrHttpOut      = "http_out"
	MetrHttpOutDescr = "HTTP out response"
)

var (
	DefaultPrometheusRegisterer = client_prometheus.DefaultRegisterer
)

// Initializes an OTLP MeterProvider
func initOtlpMeterProvider(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {

	metricExporter := must.MustVal(otlpmetricgrpc.New(ctx))

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)

	return meterProvider, nil
}

// Initializes a Prometheus MeterProvider
func initPrometheusMeterProvider(_ context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New(prometheus.WithRegisterer(DefaultPrometheusRegisterer))
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)

	return meterProvider, nil
}

// Initializes a Console MeterProvider
func initConsoleMeterProvider(res *resource.Resource, output ConsoleMeterProviderOut) (*sdkmetric.MeterProvider, error) {
	metricOptions := []stdoutmetric.Option{}
	switch output {
	case ConsoleNone:
		metricOptions = append(metricOptions, stdoutmetric.WithWriter(nullWriter{}))
	case ConsoleStderr:
		metricOptions = append(metricOptions, stdoutmetric.WithWriter(os.Stderr))
	}

	metricExporter, err := stdoutmetric.New(metricOptions...)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		// NOTE: Use the OTEL_METRIC_EXPORT_INTERVAL environment variable to control the interval
		// sdkmetric.WithInterval(3*time.Second)

		sdkmetric.WithResource(res),
	)

	return meterProvider, nil
}

func Int64CounterGetInstrument(name string, options ...metric_api.Int64CounterOption) (metric_api.Int64Counter, error) {
	return regInt64Counter.GetInstrument(name, options...)
}

func Float64CounterGetInstrument(name string, options ...metric_api.Float64CounterOption) (metric_api.Float64Counter, error) {
	return regFloat64Counter.GetInstrument(name, options...)
}

// InstrumentReg stores the already registered instruments
//
//nolint:structcheck // generics
type InstrumentReg[T any, O any] struct {
	instruments   map[string]T
	mu            sync.Mutex
	newInstrument func(name string, options ...O) (T, error)
}

// GetInstrument registers a new instrument, otherwise returns the already created.
func (r *InstrumentReg[T, O]) GetInstrument(name string, options ...O) (T, error) {
	var err error
	r.mu.Lock()
	defer r.mu.Unlock()
	instrument, has := r.instruments[name]
	if !has {
		instrument, err = r.newInstrument(name, options...)
		if err != nil {
			return instrument, fmt.Errorf("unable to register metric %T %s: %w", r, name, err)
		}
		r.instruments[name] = instrument
	}

	return instrument, nil
}

var (
	// meter is the default meter
	meter metric_api.Meter //nolint:gochecknoglobals // private
	// meterOnce is used to init meter
	meterOnce sync.Once //nolint:gochecknoglobals // private
	// regInt64Counter stores Int64Counters
	regInt64Counter *InstrumentReg[metric_api.Int64Counter, metric_api.Int64CounterOption] //nolint:gochecknoglobals // private
	// regFloat64Counter stores Float64Counters
	regFloat64Counter *InstrumentReg[metric_api.Float64Counter, metric_api.Float64CounterOption] //nolint:gochecknoglobals // private
)

// GetMeter returns the default meter.
// Inits meter and InstrumentRegs (if needed)
func GetMeter(ctx context.Context) metric_api.Meter {

	meterOnce.Do(func() {
		meter = otel.GetMeterProvider().Meter(buildinfo.ModulePath(GetMeter), metric_api.WithInstrumentationVersion("0.1"))

		regInt64Counter = &InstrumentReg[metric_api.Int64Counter, metric_api.Int64CounterOption]{
			instruments:   map[string]metric_api.Int64Counter{},
			newInstrument: meter.Int64Counter,
		}
		regFloat64Counter = &InstrumentReg[metric_api.Float64Counter, metric_api.Float64CounterOption]{
			instruments:   map[string]metric_api.Float64Counter{},
			newInstrument: meter.Float64Counter,
		}
	})

	return meter
}

func GetHost(r *http.Request) string {
	if r.Host != "" {
		return r.Host
	}

	return r.URL.Host
}
