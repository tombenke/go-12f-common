package oti

import (
	"context"
	"fmt"
	////"time"

	"github.com/tombenke/go-12f-common/must"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
)

// Initializes an OTLP MeterProvider
func initOtlpMeterProvider(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {

	metricExporter := must.MustVal(otlpmetricgrpc.New(ctx))

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}

// Initializes a Prometheus MeterProvider
func initPrometheusMeterProvider(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}

// Initializes a Console MeterProvider
func initConsoleMeterProvider(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		// NOTE: Use the OTEL_METRIC_EXPORT_INTERVAL environment variable to control the interval
		// sdkmetric.WithInterval(3*time.Second)

		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}
