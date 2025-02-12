package oti

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tombenke/go-12f-common/buildinfo"
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc/credentials/insecure"
)

type Otel struct {
	config           Config
	wg               *sync.WaitGroup
	meterProvider    *sdkmetric.MeterProvider
	tracerProvider   *sdktrace.TracerProvider
	prometheusServer *http.Server
}

// Create a Otel instance
func NewOtel(wg *sync.WaitGroup, config Config) Otel {
	return Otel{wg: wg, config: config}
}

func (o *Otel) getLogger(ctx context.Context) (context.Context, *slog.Logger) {
	return log.With(ctx, "component", "Otel")
}

// Setup the Otel providers and exporter services
func (o *Otel) Startup(ctx context.Context) {
	_, logger := o.getLogger(ctx)
	logger.Info("Starting up")

	// Create Resource for tracing and metrics
	res := must.MustVal(resource.New(ctx, resource.WithAttributes(getResourceAttributes()...)))
	res = must.MustVal(resource.Merge(res, resource.Default()))

	// Startup Metrics
	o.startupMetrics(ctx, res)

	// Startup Tracing
	o.startupTracer(ctx, res)
}

// Shut down the Otel services
func (o *Otel) Shutdown(ctx context.Context) {
	//defer o.wg.Done()
	slog.InfoContext(ctx, "Shutdown", "component", "Otel")
	o.shutdownMetrics(ctx)
	o.shutdownTracer(ctx)
}

// Startup Metrics
func (o *Otel) startupMetrics(ctx context.Context, res *resource.Resource) {
	_, logger := o.getLogger(ctx)
	exporterType := strings.ToLower(o.config.OtelMetricsExporter)
	logger.Info("Startup Metrics", "exporter", exporterType)

	switch exporterType {
	case "otlp":
		o.meterProvider = must.MustVal(initOtlpMeterProvider(ctx, res))

	case "prometheus":
		o.meterProvider = must.MustVal(initPrometheusMeterProvider(ctx, res))

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
				logger.Info("Server closed")
			} else if err != nil {
				logger.Error("Error listening for server", "error", err)
			}
			cancelCtx()
		}()

	case "console":
		o.meterProvider = must.MustVal(initConsoleMeterProvider(ctx, res))

	case "none":
		// Use no-op provider
	default:
		logger.Error("wrong metric exporter type", "otel-metric-exporter", o.config.OtelMetricsExporter)
		panic(1)
	}
}

// Shutdown Metrics
func (o *Otel) shutdownMetrics(ctx context.Context) {
	slog.InfoContext(ctx, "Shutdown", "component", "Otel.Metrics")

	// Shutdown prometheus metrics exporter server
	if o.prometheusServer != nil {
		defer o.wg.Done()
		must.Must(o.prometheusServer.Shutdown(ctx))
	}

	if o.meterProvider != nil {
		if err := o.meterProvider.Shutdown(ctx); err != nil {
			slog.ErrorContext(ctx, "failed MeterProvider shutdown", "error", err)
		}
	}
}

// Startup Tracer
func (o *Otel) startupTracer(ctx context.Context, res *resource.Resource) {
	_, logger := o.getLogger(ctx)

	exporterType := strings.ToLower(o.config.OtelTracesExporter)
	logger.Info("Startup Tracing", "exporter", exporterType)

	switch exporterType {
	case "otlp":
		o.tracerProvider = must.MustVal(initOtlpTracerProvider(ctx, res, o.config.OtelTracesSampler, o.config.OtelTracesSamplerArg))
	/*
		case "jaeger":
			o.tracerProvider = must.MustVal(initJaegerTracerProvider(ctx, res, o.config.OtelTracesSampler, o.config.OtelTracesSamplerArg))

		case "zipkin":
			o.tracerProvider = must.MustVal(initZipkinTracerProvider(ctx, res, o.config.OtelTracesSampler, o.config.OtelTracesSamplerArg))
	*/
	case "console":
		o.tracerProvider = must.MustVal(initConsoleTracerProvider(ctx, res, o.config.OtelTracesSampler, o.config.OtelTracesSamplerArg))

	case "none":
		// Use no-op provider
	default:
		logger.Error("wrong tracer exporter type", "otel-traces-exporter", o.config.OtelTracesExporter)
		panic(1)
	}
}

// Shutdown Tracing
func (o *Otel) shutdownTracer(ctx context.Context) {
	slog.InfoContext(ctx, "Shutdown", "component", "Otel.Tracer")

	if o.tracerProvider != nil {
		if err := o.tracerProvider.Shutdown(ctx); err != nil {
			slog.ErrorContext(ctx, "failed TracerProvider shutdown", "error", err)
		}
	}
}

// Initialize a gRPC connection to be used by both the tracer and meter providers.
func initOtelGrpcConn(ctx context.Context) (*grpc.ClientConn, error) {
	// It connects the OpenTelemetry Collector through local gRPC connection.
	// TODO: Replace `localhost:4317` with config parameters
	otelCollectorUrl := "localhost:4317"
	slog.InfoContext(ctx, "Connect to OTEL Collector", "url", otelCollectorUrl)
	conn, err := grpc.NewClient(otelCollectorUrl,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

func getResourceAttributes() []attribute.KeyValue {
	attributes := []attribute.KeyValue{}

	// Add service.version attribute if it is defined
	if buildinfo.Version() != "" {
		attributes = append(attributes, semconv.ServiceVersionKey.String(buildinfo.Version()))
	}

	// TODO: May add the following attributes
	// semconv.ServiceInstanceIDKey.String("this-is-a-service-instance-ID"),
	// ???

	return attributes
}
