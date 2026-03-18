package oti

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/v2/log"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TestOtelStartupWithEmptyExporters tests the Otel.Startup method when both
// OtelMetricsExporter and OtelTracesExporter are set to empty strings.
// This should result in no-op providers being set up.
func TestOtelStartupWithEmptyExporters(t *testing.T) {
	// Arrange
	config := Config{
		OtelMetricsExporter:        "",
		OtelTracesExporter:         "",
		OtelExporterPrometheusPort: 0,
	}

	wg := &sync.WaitGroup{}
	otelInstance := NewOtel(wg, config)

	ctx := context.Background()

	// Act - should not panic
	resultCtx := otelInstance.Startup(ctx)

	// Assert - Context
	assert.NotNil(t, resultCtx, "Startup should return a non-nil context")

	// Log test - should have a logger in the context even with no exporters

	// Assert - Logger from context
	ctxLogger, logger := log.FromContext(resultCtx)
	assert.NotNil(t, ctxLogger, "Logger context should not be nil")
	assert.NotNil(t, logger, "Logger should not be nil")
	logger.Info("Logger is available in context even with no exporters")

	// Metric test - should be a no-op provider

	// Assert - MeterProvider
	meterProvider := otel.GetMeterProvider()
	assert.NotNil(t, meterProvider, "MeterProvider should be set")

	// Check if it's an SDK MeterProvider (when empty exporter, it should create a no-op one)
	sdkMeterProvider, isSdkMeter := meterProvider.(*sdkmetric.MeterProvider)
	assert.True(t, isSdkMeter, "MeterProvider should be an SDK MeterProvider")

	// Verify we can create a meter from the provider
	meter := sdkMeterProvider.Meter("test-meter")
	assert.NotNil(t, meter, "Should be able to create a meter from the provider")

	// Trace test - should be a no-op provider

	// Assert - TracerProvider
	tracerProvider := otel.GetTracerProvider()
	assert.NotNil(t, tracerProvider, "TracerProvider should be set")

	// Check if it's an SDK TracerProvider (when empty exporter, it should create a no-op one)
	sdkTracerProvider, isSdkTracer := tracerProvider.(*sdktrace.TracerProvider)
	assert.True(t, isSdkTracer, "TracerProvider should be an SDK TracerProvider")

	// Verify we can create a tracer from the provider
	tracer := sdkTracerProvider.Tracer("test-tracer")
	assert.NotNil(t, tracer, "Should be able to create a tracer from the provider")

	// Cleanup
	otelInstance.Shutdown(ctx)
}
