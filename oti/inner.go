package oti

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	metric_api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type InternalMiddlewareFn[T any] func(ctx context.Context) (T, error)

type InternalMiddleware[T any] func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T]

func InternalMiddlewareChain[T any](mws ...InternalMiddleware[T]) InternalMiddleware[T] {
	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		fn := next
		for mw := len(mws) - 1; mw >= 0; mw-- {
			fn = mws[mw](fn)
		}

		return fn
	}
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

	return strings.SplitN(err.Error(), ":", 2)[0]
}

// Matches from the start until the first occurrence of ';', ':', or ','
// (i.e. the characters before the first of those punctuation marks)
var firstErrPartRe = regexp.MustCompile(`^[^;:,]+`)

// FirstErrPart returns the characters before first ;:, characters
func FirstErrPart(err error) string {
	if err == nil {
		return ""
	}

	return firstErrPartRe.FindString(err.Error())
}

var (
	// ErrTypeCast is an error for type assertion from interface
	ErrTypeCast     = errors.New("unable to cast interface to type")
	ErrorEventName  = "ERROR"
	ErrUnableToCast = errors.New("interface cast error")
)

// Span is a middleware to start/end a new span, using from context.
// Sets "traceID", "spanParentID" and "spanID" log values.
func Span[T any](tr trace.Tracer, spanKind trace.SpanKind, spanName string) InternalMiddleware[T] {
	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		return func(ctx context.Context) (T, error) {
			spanParent := trace.SpanFromContext(ctx).SpanContext()
			ctx, spanChild := tr.Start(ctx, spanName,
				trace.WithSpanKind(spanKind),
			)
			defer spanChild.End()

			logFields := []any{}
			if !spanParent.IsValid() {
				logFields = append(logFields, FieldTraceID, spanChild.SpanContext().TraceID().String())
			}
			logFields = append(logFields, FieldSpanID, spanChild.SpanContext().SpanID().String())
			ctx = LogWithValues(ctx, logFields...)

			t, err := next(ctx)
			if err != nil {
				spanChild.RecordError(err, trace.WithAttributes(FieldValue.String(fmt.Sprintf("%+v", t))))
				spanChild.SetStatus(codes.Error, err.Error())
			} else {
				spanChild.AddEvent(EventOK, trace.WithAttributes(FieldValue.String(fmt.Sprintf("%+v", t))))
			}
			return t, err
		}
	}
}

// TryCatch is a middleware for catching Go panic and propagating it as an error
func TryCatch[T any]() InternalMiddleware[T] {
	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		return func(ctx context.Context) (T, error) {
			var retVal T
			var err error
			if errTryCatch := tryCatch(func() {
				retVal, err = next(ctx)
			})(); errTryCatch != nil {
				err = errTryCatch
			}

			return retVal, err
		}
	}
}

// ErrPanic is an error for captured panic
var ErrPanic = errors.New("captured panic")

// tryCatch captures a Go panic and returns as an error
func tryCatch(f func()) func() error {
	return func() (err error) {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				err = fmt.Errorf("%w: %v, %s", ErrPanic, panicInfo, string(debug.Stack()))

				return
			}
		}()

		f() // calling the decorated function

		return err
	}
}

// Logger is a middleware for logging begin and end messages.
// A new logger with values is added to the context.
func Logger[T any](
	beginMessage string, beginLevel int, endMessage string, endLevel int, errMessage string,
	values []attribute.KeyValue,
) InternalMiddleware[T] {
	logValues := make([]any, 0, len(values)*2)
	for _, v := range values {
		logValues = append(logValues, v.Key, v.Value.Emit())
	}

	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		return func(ctx context.Context) (T, error) {
			var err error
			ctx = LogWithValues(ctx, logValues...)
			Log(ctx, beginLevel, beginMessage)
			beginTS := time.Now()

			retVal, err := next(ctx)

			elapsedSec := time.Since(beginTS).Seconds()
			args := []any{FieldDuration, fmt.Sprintf("%.3f", elapsedSec)}
			if err != nil {
				LogError(ctx, err, errMessage, args...)
			} else {
				Log(ctx, endLevel, endMessage, args...)
			}

			return retVal, err
		}
	}
}

/*
Metrics is a middleware to make count and duration report

	Prometheus-specific implementation:
	The "_total" suffix is appended to the counter name, defined in "counterSuffix", see:
	https://github.com/open-telemetry/opentelemetry-go/blob/main/exporters/prometheus/exporter.go#L100
	The unit "s" is appended as "_seconds" to the metric name (injected before the "_total" suffix),
	defined in "unitSuffixes", see
	https://github.com/open-telemetry/opentelemetry-go/blob/main/exporters/prometheus/exporter.go#L343
*/
func Metrics[T any](ctx context.Context, meter metric_api.Meter, name string,
	description string, attributes []attribute.KeyValue, errFormatter ErrFormatter,
) InternalMiddleware[T] {
	// Copy, because error will be added to the end of the slice

	attempted, err := Int64CounterGetInstrument(name, metric_api.WithDescription(description))
	if err != nil {
		LogError(ctx, err, "unable to instantiate counter", FieldMetricName, name)
		panic(err)
	}
	durationSum, err := Float64CounterGetInstrument(name+"_duration", metric_api.WithDescription(description+", duration sum"), metric_api.WithUnit("s"))
	if err != nil {
		LogError(ctx, err, "unable to instantiate time counter", FieldMetricName, name)
		panic(err)
	}

	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		return func(ctx context.Context) (T, error) {
			beginTS := time.Now()

			retVal, err := next(ctx)

			elapsedSec := time.Since(beginTS).Seconds()
			// To be thread-safe
			baseAttrs := make([]attribute.KeyValue, len(attributes), len(attributes)+1)
			copy(baseAttrs, attributes)
			opt := metric_api.WithAttributes(append(baseAttrs, MetrAttrErr.String(errFormatter(err)))...)
			attempted.Add(ctx, 1, opt)
			durationSum.Add(ctx, elapsedSec, opt)

			return retVal, err
		}
	}
}
