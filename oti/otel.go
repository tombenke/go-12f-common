package oti

import (
	"context"
	"errors"
	"fmt"
	////"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tombenke/go-12f-common/v2/buildinfo"
	"github.com/tombenke/go-12f-common/v2/log"
	"github.com/tombenke/go-12f-common/v2/must"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Otel struct {
	config           Config
	wg               *sync.WaitGroup
	prometheusServer *http.Server
}

// nullWriter implements io.Writer and discards all data written to it.
type nullWriter struct{}

// Write implements io.Writer for nullWriter.
func (nullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

type ConsoleMeterProviderOut int

const (
	ConsoleNone ConsoleMeterProviderOut = iota
	ConsoleStdout
	ConsoleStderr
)

type MetricExporterType string

var (
	MetricExporterTypeOTLP       MetricExporterType = "otlp"
	MetricExporterTypePrometheus MetricExporterType = "prometheus"
	MetricExporterTypeConsole    MetricExporterType = "console"
	MetricExporterTypeNone       MetricExporterType = "none"
)

type TraceExporterType string

var (
	TraceExporterTypeOTLP    TraceExporterType = "otlp"
	TraceExporterTypeConsole TraceExporterType = "console"
	TraceExporterTypeNone    TraceExporterType = "none"
)

// Create a Otel instance
func NewOtel(wg *sync.WaitGroup, config Config) Otel {
	return Otel{wg: wg, config: config}
}

////func (o *Otel) getLogger(ctx context.Context) (context.Context, *slog.Logger) {
////	return log.With(ctx, FieldComponent, "Otel")
////}

// Setup the Otel providers and exporter services
func (o *Otel) Startup(ctx context.Context) context.Context {
	resAttrs := getResourceAttributes()
	resFields := []any{}
	for _, attr := range resAttrs {
		resFields = append(resFields, string(attr.Key), attr.Value.AsString())
	}
	ctx = LogWithValues(ctx, resFields...)
	ctxLog, _ := log.FromContext(ctx, string(FieldComponent), "Otel")
	Log(ctxLog, 0, "Starting up")

	// Create Resource for tracing and metrics
	res := must.MustVal(resource.New(ctx))
	res = must.MustVal(resource.Merge(res, resource.Default()))
	res = must.MustVal(resource.Merge(res, must.MustVal(resource.New(ctx, resource.WithAttributes(getResourceAttributes()...)))))

	// Startup Metrics
	o.startupMetrics(ctx, res)

	// Startup Tracing
	o.startupTracer(ctx, res)

	return ctx
}

// Shut down the Otel services
func (o *Otel) Shutdown(ctx context.Context) {
	//defer o.wg.Done()
	ctx = LogWithValues(ctx, FieldComponent, "Otel")

	Log(ctx, 0, "Shutdown")
	o.shutdownMetrics(ctx)
	o.shutdownTracer(ctx)
}

// Startup Metrics
func (o *Otel) startupMetrics(ctx context.Context, res *resource.Resource) {
	exporterType := strings.ToLower(o.config.OtelMetricsExporter)
	Log(ctx, 0, "Startup Metrics", FieldMetricExporter, exporterType)
	var meterProvider *sdkmetric.MeterProvider

	switch MetricExporterType(exporterType) {
	case MetricExporterTypeOTLP:
		meterProvider = must.MustVal(initOtlpMeterProvider(ctx, res))

	case MetricExporterTypePrometheus:
		meterProvider = must.MustVal(initPrometheusMeterProvider(ctx, res))

		if o.config.OtelExporterPrometheusPort > 0 {
			_, cancelCtx := context.WithCancel(context.Background())
			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.Handler())
			o.prometheusServer = &http.Server{
				Addr:    fmt.Sprintf(":%d", o.config.OtelExporterPrometheusPort),
				Handler: mux,
			}

			// Start the blocking server call in a separate thread
			o.wg.Add(1)
			go func() {
				// Start prometheus metrics exporter server
				err := o.prometheusServer.ListenAndServe()
				if errors.Is(err, http.ErrServerClosed) {
					Log(ctx, 0, "Server closed")
				} else if err != nil {
					LogError(ctx, err, "Error listening for server")
				}
				cancelCtx()
			}()
		}

	case MetricExporterTypeConsole:
		meterProvider = must.MustVal(initConsoleMeterProvider(res, ConsoleStdout))

	case MetricExporterTypeNone:
		// Use no-op provider
		meterProvider = must.MustVal(initConsoleMeterProvider(res, ConsoleNone))
	default:
		LogError(ctx, ErrOtelConfig, "wrong metric exporter type", "otel-metric-exporter", o.config.OtelMetricsExporter)
		panic(1)
	}

	otel.SetMeterProvider(meterProvider)
}

// Shutdown Metrics
func (o *Otel) shutdownMetrics(ctx context.Context) {
	ctx = LogWithValues(ctx, FieldComponent, "Otel.Metrics")
	Log(ctx, 0, "Shutdown")

	// Shutdown prometheus metrics exporter server
	if o.prometheusServer != nil {
		defer o.wg.Done()
		must.Must(o.prometheusServer.Shutdown(ctx))
	}

	if meterProvider, is := otel.GetMeterProvider().(*sdkmetric.MeterProvider); is {
		if err := meterProvider.Shutdown(ctx); err != nil {
			LogError(ctx, err, "failed MeterProvider shutdown")
		}
	}
}

// Startup Tracer
func (o *Otel) startupTracer(ctx context.Context, res *resource.Resource) context.Context {
	exporterType := strings.ToLower(o.config.OtelTracesExporter)
	Log(ctx, 0, "Startup Tracing", FieldExporter, exporterType)

	var tracerProvider *sdktrace.TracerProvider
	switch TraceExporterType(exporterType) {
	case TraceExporterTypeOTLP:
		tracerProvider = must.MustVal(initTracerProvider(ctx, must.MustVal(otlptracegrpc.New(ctx)), res))

		/*
			case "jaeger":
				tracerProvider = must.MustVal(initJaegerTracerProvider(ctx, res))

			case "zipkin":
				tracerProvider = must.MustVal(initZipkinTracerProvider(ctx, res))
		*/

	case TraceExporterTypeConsole:
		tracerProvider = must.MustVal(initTracerProvider(ctx, must.MustVal(stdouttrace.New(stdouttrace.WithPrettyPrint())), res))

	case TraceExporterTypeNone, "":
		// Use no-op provider
		tracerProvider = must.MustVal(initTracerProvider(ctx, must.MustVal(stdouttrace.New(stdouttrace.WithWriter(nullWriter{}))), res))
	default:
		LogError(ctx, ErrOtelConfig, "wrong tracer exporter type", "otel-traces-exporter", o.config.OtelTracesExporter)
		panic(1)
	}

	otel.SetTracerProvider(tracerProvider)

	return ctx
}

// Shutdown Tracing
func (o *Otel) shutdownTracer(ctx context.Context) {
	ctx = LogWithValues(ctx, FieldComponent, "Otel.Tracer")
	Log(ctx, 0, "Shutdown")

	if tracerProvider, is := otel.GetTracerProvider().(*sdktrace.TracerProvider); is {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			LogError(ctx, err, "failed TracerProvider shutdown")
		}
	}
}

// getResourceAttributes returns common resource attributes for Otel telemetry data.
// Shall be used for both logs, tracing and metrics.
// Check conventions https://helm.sh/docs/chart_best_practices/labels/ .
// If it's set via Helm labels, these attributes can be skipped here to avoid duplication.
func getResourceAttributes() []attribute.KeyValue {
	hostname, _ := os.Hostname() //nolint:errcheck // not important
	podNamespace := os.Getenv("POD_NAMESPACE")
	attributes := []attribute.KeyValue{
		semconv.ServiceNamespaceKey.String(podNamespace),
		FieldApp.String(buildinfo.AppName()),
		semconv.ServiceInstanceIDKey.String(hostname),
		semconv.ServiceVersionKey.String(buildinfo.Version()),
	}

	// TODO: May add further attributes here

	// semconv.ServiceInstanceIDKey.String("this-is-a-service-instance-ID"),
	// ???

	return attributes
}
