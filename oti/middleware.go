package oti

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	metric_api "go.opentelemetry.io/otel/metric"

	logger "github.com/tombenke/go-12f-common/v2/log"
)

type InternalMiddlewareFn[T any] func(ctx context.Context) (T, error)

type InternalMiddleware[T any] func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T]

type ComponentNamer interface {
	ComponentName() string
}

// ErrFormatter is a func type to format metric error attribute
type ErrFormatter func(error) string

// NoErr always returns "". Can be used to skip any error stats in the metrics
func NoErr(error) string {
	return ""
}

// FullErr returns the full error text.
// Be careful about the cardinality, if the error text has dynamic part(s) (see: Prometheus label)
func FullErr(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

// FirstErr returns the first part of error text before ':'
func FirstErr(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()
	if i := strings.Index(errMsg, ":"); i < 0 {
		return errMsg
	} else {
		return errMsg[:i]
	}
}

var (
	ErrUnableToCast = errors.New("interface cast error")
	// ErrTypeCast is an error for type assertion from interface
	ErrTypeCast    = errors.New("unable to cast interface to type")
	ErrorEventName = "ERROR"
)

/*
Metrics is a middleware to make count and duration report

	Prometheus-specific implementation:
	The "_total" suffix is appended to the counter name, defined in "counterSuffix", see:
	https://github.com/open-telemetry/opentelemetry-go/blob/main/exporters/prometheus/exporter.go#L100
	The unit "s" is appended as "_seconds" to the metric name (injected before the "_total" suffix),
	defined in "unitSuffixes", see
	https://github.com/open-telemetry/opentelemetry-go/blob/main/exporters/prometheus/exporter.go#L343
*/
func Metrics[T any](ctx context.Context, name string,
	description string, attributes map[string]string, errFormatter ErrFormatter,
) InternalMiddleware[T] {
	log := logger.GetFromContextOrDefault(ctx)
	baseAttrs := make([]attribute.KeyValue, 0, len(attributes))
	for aKey, aVal := range attributes {
		baseAttrs = append(baseAttrs, attribute.Key(aKey).String(aVal))
	}
	attempted, err := Int64CounterGetInstrument(name, metric_api.WithDescription(description))
	if err != nil {
		log.Error("unable to instantiate counter", KeyError, err, "metricName", name)
		panic(err)
	}
	durationSum, err := Float64CounterGetInstrument(name+"_duration", metric_api.WithDescription(description+", duration sum"), metric_api.WithUnit("s"))
	if err != nil {
		log.Error("unable to instantiate time counter", KeyError, err, "metricName", name)
		panic(err)
	}

	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		return func(ctx context.Context) (T, error) {
			beginTS := time.Now()

			retVal, err := next(ctx)

			elapsedSec := time.Since(beginTS).Seconds()
			attrs := make([]attribute.KeyValue, len(baseAttrs), len(baseAttrs)+1)
			copy(attrs, baseAttrs)
			opt := metric_api.WithAttributes(append(attrs, attribute.Key(MetrAttrErr).String(errFormatter(err)))...)
			attempted.Add(ctx, 1, opt)
			durationSum.Add(ctx, elapsedSec, opt)

			return retVal, err
		}
	}
}
