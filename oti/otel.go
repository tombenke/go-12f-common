package oti

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/tombenke/go-12f-common/log"
	//"go.opentelemetry.io/otel/sdk/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Otel struct {
	config         Config
	wg             *sync.WaitGroup
	meterProvider  *sdkmetric.MeterProvider
	tracerProvider *sdktrace.TracerProvider
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

	// TODO: Create Resource for tracing and metrics
	//res, err := resource.New(ctx,
	//	resource.WithAttributes(
	//		// The service name used to display traces in backends
	//		serviceName,
	//	),
	//)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// Startup Metrics
	o.startupMetrics(ctx)

	// Startup Tracing
	o.startupTracer(ctx)

	// TODO: Setup prometheus metrics exporter server
	_, cancelCtx := context.WithCancel(context.Background())

	// Start the blocking server call in a separate thread
	o.wg.Add(1)
	go func() {
		// TODO: Start prometheus metrics exporter server
		cancelCtx()
	}()
}

// Shut down the Otel services
func (o *Otel) Shutdown(ctx context.Context) {
	defer o.wg.Done()
	slog.InfoContext(ctx, "Shutdown", "component", "Otel")
	o.shutdownMetrics(ctx)
	o.shutdownTracer(ctx)
}

// Startup Metrics
func (o *Otel) startupMetrics(ctx context.Context) {
	_, logger := o.getLogger(ctx)
	exporterType := strings.ToLower(o.config.OtelMetricsExporter)
	logger.Info("Startup Metrics", "exporter", exporterType)

	switch exporterType {
	case "otlp":
		conn, connErr := initOtelGrpcConn(ctx) // TODO: use it only in case of tracer or metrics uses otel exporter
		if connErr != nil {
			logger.Error("failed to create grpc connection", "error", connErr)
			panic(1)
		}
		o.meterProvider, _ = initOtlpMeterProvider(ctx, conn)

	case "prometheus":
		o.meterProvider, _ = initPrometheusMeterProvider(ctx)

	case "console":
		o.meterProvider, _ = initConsoleMeterProvider(ctx)

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

	// TODO: Shutdown prometheus metrics exporter server
	//must.Must(o.server.Shutdown(context.Background()))

	if o.meterProvider != nil {
		if err := o.meterProvider.Shutdown(ctx); err != nil {
			slog.ErrorContext(ctx, "failed MeterProvider shutdown", "error", err)
		}
	}
}

// Startup Tracer
func (o *Otel) startupTracer(ctx context.Context) {
	_, logger := o.getLogger(ctx)
	logger.Info("Startup Tracer")
	// TODO
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
