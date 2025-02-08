package oti

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	//"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/prometheus"
	//semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

// Initializes an OTLP MeterProvider
func initOtlpMeterProvider(ctx context.Context /*, res *resource.Resource*/, conn *grpc.ClientConn) (*sdkmetric.MeterProvider, error) {

	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))

	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		//sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}

// Initializes a Prometheus MeterProvider
func initPrometheusMeterProvider(ctx context.Context /*, res *resource.Resource*/) (*sdkmetric.MeterProvider, error) {
	// TODO
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}

// Initializes a Console MeterProvider
func initConsoleMeterProvider(ctx context.Context /*, res *resource.Resource*/) (*sdkmetric.MeterProvider, error) {
	// TODO
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			sdkmetric.WithInterval(3*time.Second))),
	)

	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}
