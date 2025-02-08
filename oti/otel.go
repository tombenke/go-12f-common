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
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

}

// Shut down the Otel services
func (o *Otel) Shutdown(ctx context.Context) {
	//defer o.wg.Done()
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

		// TODO: Handle error as fatal
		o.meterProvider, _ = initOtlpMeterProvider(ctx, conn)

	case "prometheus":
		// TODO: Handle error as fatal
		o.meterProvider, _ = initPrometheusMeterProvider(ctx)

		// TODO: Setup prometheus metrics exporter server
		_, cancelCtx := context.WithCancel(context.Background())
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		o.prometheusServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", 2223 /*o.config.PrometheusPort*/),
			Handler: mux,
		}

		// Start the blocking server call in a separate thread
		o.wg.Add(1)
		go func() {
			// TODO: Start prometheus metrics exporter server
			err := o.prometheusServer.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				logger.Info("Server closed")
			} else if err != nil {
				logger.Error("Error listening for server", "error", err)
			}
			cancelCtx()
		}()

	case "console":
		// TODO: Handle error as fatal
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

// Initialize a gRPC connection to be used by both the tracer and meter providers.
func initOtelGrpcConn(ctx context.Context) (*grpc.ClientConn, error) {
	// It connects the OpenTelemetry Collector through local gRPC connection.
	// TODO: Replace `localhost:4317` with config parameter
	conn, err := grpc.NewClient("localhost:4317",
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}
