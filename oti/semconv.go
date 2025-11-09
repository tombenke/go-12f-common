package oti

/*
This file contains similar keys to OTEL semconv.

OTEL naming rules: https://opentelemetry.io/docs/specs/semconv/general/naming
    If a name is a tree node (namespace), it cannot be leaf. For example:
    If "error" is leaf, which contains the error message, the "error.type" cannot be used.
    The "error" should be a tree node (namespace), the "error.message" and "error.type" can be leaf.
Latest OTEL semconv spec release: https://opentelemetry.io/docs/specs/semconv/
OTEL semconv Go releases: https://github.com/open-telemetry/opentelemetry-go/tree/main/semconv

Similar to https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels,
Mimir does not accept '.' in metric labels, which is replaced with '_' by OTEL libraries.

Loki does not accept '.' in label filters, which is replaced with '_' by Telemetry Gateway.
List of label filters must be configured in Telemetry Gateway config (logs.*.otelCustomConfig).
*/

const (
	// Log fields
	KeyError      = "error.message"
	KeyMetricName = "metric.name"
	KeyTestSuite  = "test.suite"

	// Metric labels
	KeyApp     = "app"
	KeyService = "service"
)
