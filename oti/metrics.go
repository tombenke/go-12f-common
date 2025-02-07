package oti

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	////"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	//"go.opentelemetry.io/otel/sdk/resource"
	//semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
	return nil, nil
	//return meterProvider, nil
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
