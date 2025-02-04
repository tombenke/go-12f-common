// Goroutine middlewares
package inner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	metric_api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/semaphore"

	logger "github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/middleware"
)

var (
	// ErrTypeCast is an error for type assertion from interface
	ErrTypeCast    = errors.New("unable to cast interface to type")
	ErrorEventName = "ERROR"
)

// SemAcquire is a middleware to acquire semaphore
func SemAcquire[T any](sem *semaphore.Weighted) InternalMiddleware[T] {
	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		return func(ctx context.Context) (T, error) {
			if err := sem.Acquire(ctx, 1); err != nil {
				err = fmt.Errorf("cannot acquire semaphore: %w", err)
				var zero T

				return zero, err
			}
			defer func() {
				select {
				case <-ctx.Done():
				default:
					sem.Release(1)
				}
			}()

			return next(ctx)
		}
	}
}

// Span is a middleware to start/end a new span, using from context.
// Sets "traceID", "spanParentID" and "spanID" log values.
func Span[T any](tr trace.Tracer, spanName string) InternalMiddleware[T] {
	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		return func(ctx context.Context) (T, error) {
			spanParent := trace.SpanFromContext(ctx).SpanContext()
			spanKind := trace.SpanKindInternal
			ctx, spanChild := tr.Start(ctx, spanName,
				trace.WithSpanKind(spanKind),
			)
			defer spanChild.End()

			logFields := []any{}
			if !spanParent.IsValid() {
				logFields = append(logFields, "traceID", spanChild.SpanContext().TraceID().String())
			}
			logFields = append(logFields, "spanID", spanChild.SpanContext().SpanID().String())
			ctx, _ = logger.FromContext(ctx, logFields...)

			t, err := next(ctx)
			if err != nil {
				spanChild.RecordError(err, trace.WithAttributes(attribute.String("value", fmt.Sprintf("%+v", t))))
				spanChild.SetStatus(codes.Error, err.Error())
			} else {
				spanChild.AddEvent("ok", trace.WithAttributes(attribute.String("value", fmt.Sprintf("%+v", t))))
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
func Logger[T any](values map[string]string, beginLevel slog.Level, endLevel slog.Level) InternalMiddleware[T] {
	logValues := make([]any, 0, len(values)*2)
	for k, v := range values {
		logValues = append(logValues, k, v)
	}

	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		return func(ctx context.Context) (T, error) {
			var log *slog.Logger
			var err error
			ctx, log = logger.FromContext(ctx, logValues...)
			log.Log(ctx, beginLevel, "INT_BEGIN")
			beginTS := time.Now()
			defer func() {
				elapsedSec := time.Since(beginTS).Seconds()
				args := []any{"duration", fmt.Sprintf("%.3f", elapsedSec)}
				if err != nil {
					args = append(args, logger.KeyError, err)
				}
				log.With(args...).Log(ctx, endLevel, "INT_END")
			}()

			retVal, err := next(ctx)

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
	description string, attributes map[string]string, errFormatter middleware.ErrFormatter,
) InternalMiddleware[T] {
	_, log := logger.FromContext(ctx)
	baseAttrs := make([]attribute.KeyValue, 0, len(attributes))
	for aKey, aVal := range attributes {
		baseAttrs = append(baseAttrs, attribute.Key(aKey).String(aVal))
	}
	attempted, err := middleware.Int64CounterGetInstrument(name, metric_api.WithDescription(description))
	if err != nil {
		log.Error("unable to instantiate counter", logger.KeyError, err, "metricName", name)
		panic(err)
	}
	durationSum, err := middleware.Float64CounterGetInstrument(name+"_duration", metric_api.WithDescription(description+", duration sum"), metric_api.WithUnit("s"))
	if err != nil {
		log.Error("unable to instantiate time counter", logger.KeyError, err, "metricName", name)
		panic(err)
	}

	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		return func(ctx context.Context) (T, error) {
			beginTS := time.Now()

			retVal, err := next(ctx)

			elapsedSec := time.Since(beginTS).Seconds()
			attrs := make([]attribute.KeyValue, len(baseAttrs), len(baseAttrs)+1)
			copy(attrs, baseAttrs)
			opt := metric_api.WithAttributes(append(attrs, attribute.Key(middleware.MetrAttrErr).String(errFormatter(err)))...)
			attempted.Add(ctx, 1, opt)
			durationSum.Add(ctx, elapsedSec, opt)

			return retVal, err
		}
	}
}
