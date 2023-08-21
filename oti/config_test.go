package oti_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/oti"
)

func TestNewConfigWithDefaults(t *testing.T) {
	assert.Equal(t, &oti.Config{
		ServiceName: oti.DefaultServiceName,
		Exporter: oti.ExporterConfig{
			Type:         oti.DefaultExporterType,
			CollectorURL: oti.DefaultCollectorURL,
		},
		Sampling: oti.SamplingConfig{
			Type:  oti.DefaultTraceSamplingType,
			Ratio: -1,
		},
		SpanProcessorType: oti.DefaultSpanProcessorType,
	}, oti.NewConfig())
}

func TestNewConfigWithOptions(t *testing.T) {
	serviceName := "test-service"
	exporterType := oti.StdoutExporter

	assert.Equal(t, &oti.Config{
		ServiceName: oti.ServiceName(serviceName),
		Exporter: oti.ExporterConfig{
			Type:         exporterType,
			CollectorURL: "",
		},
		Sampling: oti.SamplingConfig{
			Type:  oti.RatioBasedSampling,
			Ratio: 0.001,
		},
		SpanProcessorType: oti.BatchSpanProcessor,
	}, oti.NewConfig(
		oti.WithServiceName(serviceName),
		oti.WithStdoutExporter(),
		oti.WithRatioBasedSampling(0.001),
		oti.WithBatchSpanProcessor(),
	))
}

func TestWithExporter(t *testing.T) {
	type ETC struct {
		Param    string
		Expected oti.ExporterType
	}

	// Check valid cases
	for _, testCase := range []ETC{
		{Param: "Stdout", Expected: oti.StdoutExporter},
		{Param: "STDOUT", Expected: oti.StdoutExporter},
		{Param: "OTELGRPC", Expected: oti.OtelGrpcExporter},
	} {
		result, err := oti.WithExporter(testCase.Param)
		assert.Nil(t, err)
		assert.Equal(t, testCase.Expected, result)
	}

	// Check invalid case
	result, err := oti.WithExporter("Invalid")
	assert.NotNil(t, err)
	assert.Equal(t, oti.DefaultExporterType, result)
}

func TestWithSpanProcessor(t *testing.T) {
	type ETC struct {
		Param    string
		Expected oti.SpanProcessorType
	}

	// Check valid cases
	for _, testCase := range []ETC{
		{Param: "SIMPLE", Expected: oti.SimpleSpanProcessor},
		{Param: "BATCH", Expected: oti.BatchSpanProcessor},
	} {
		result, err := oti.WithSpanProcessor(testCase.Param)
		assert.Nil(t, err)
		assert.Equal(t, testCase.Expected, result)
	}

	// Check invalid case
	result, err := oti.WithSpanProcessor("Invalid")
	assert.NotNil(t, err)
	assert.Equal(t, oti.DefaultSpanProcessorType, result)
}

func TestWithTraceSampling(t *testing.T) {
	type ETC struct {
		Param    string
		Expected oti.TraceSamplingType
	}

	// Check valid cases
	for _, testCase := range []ETC{
		{Param: "NEVER", Expected: oti.NeverSampling},
		{Param: "PARENTBASED", Expected: oti.ParentBasedSampling},
		{Param: "RATIOBASED", Expected: oti.RatioBasedSampling},
		{Param: "ALWAYS", Expected: oti.AlwaysSampling},
	} {
		result, err := oti.WithTraceSampling(testCase.Param)
		assert.Nil(t, err)
		assert.Equal(t, testCase.Expected, result)
	}

	// Check invalid case
	result, err := oti.WithTraceSampling("Invalid")
	assert.NotNil(t, err)
	assert.Equal(t, oti.DefaultTraceSamplingType, result)
}

// ExampleConfig demonstrates how to initialize and use the OTEL instrumentation in your application
func ExampleConfig() {
	// Create a new OTEL configuration
	otelConfig := *oti.NewConfig(
		// Define the name of your service
		oti.WithServiceName("MPA-Test"),

		// Select an exporter
		oti.WithStdoutExporter(),
		//oti.WithOtelGrpcCollectorExporter("localhost:4317"),

		// Set the type of sampling
		oti.WithNeverSampling(),
		//oti.WithAlwaysSampling(),

		// Select the type of span processor
		oti.WithSimpleSpanProcessor(),
		oti.WithBatchSpanProcessor(),
	)

	// Initialize OTEL instrumentation using the config
	appOT := oti.NewOTI(otelConfig)

	// Use the OTEL trace and metrics functions according your need
	// in case you are using libraries with built-in instrumentation e.g. RPC or MPA then
	// you do not have to add more OTEL related code,
	// except you want to extend it with your own spans or metrics

	// Shuts down the OTEL instrumentation
	defer appOT.Shutdown()
}
