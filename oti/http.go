package oti

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	metric_api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	// latest semconv version which supports HTTP*AttributesFromHTTPRequest
	semconv_legacy "go.opentelemetry.io/otel/semconv/v1.12.0"
)

/*
 Server HTTP middleware
*/

type HttpMiddlewareFunc func(http.Handler) http.Handler

// HttpServerLoggerBaseMiddleware sets a minimal logger to the request context,
// before the Span middleware
func HttpServerLoggerBaseMiddleware(ctxRoot context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var urlFull, urlHost, urlPath string
			if r.URL != nil {
				urlFull = r.URL.String()
				urlHost = r.URL.Host
				urlPath = r.URL.Path
			}
			ctx := CopyLogger(r.Context(), LogWithValues(ctxRoot,
				FieldHttpMethod, r.Method,
				FieldUrlFull, urlFull,
				FieldUrlHost, urlHost,
				FieldUrlPath, urlPath,
			))

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

// TODO TraceState propagation (client command)
// spanLogLevel: Debug
func HttpServerTracerMiddleware(tr trace.Tracer, instance string, spanLogLevel int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			routePath := getRoutePath(r)

			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))
			span := trace.SpanFromContext(ctx)
			if span.SpanContext().IsValid() {
				spanValues, spanValuesErr := span.SpanContext().MarshalJSON()
				Log(ctx, spanLogLevel, MsgSpanIn,
					FieldSpan, spanValues,
					FieldSpanErr, spanValuesErr,
					FieldTraceID, span.SpanContext().TraceID().String(),
					FieldSpanID, span.SpanContext().SpanID().String(),
				)
			} else {
				span = trace.SpanFromContext(ctx)
				spanValues, spanValuesErr := span.SpanContext().MarshalJSON()
				Log(ctx, spanLogLevel, MsgSpanNew,
					FieldSpan, spanValues,
					FieldSpanErr, spanValuesErr,
					FieldTraceID, span.SpanContext().TraceID().String(),
					FieldSpanID, span.SpanContext().SpanID().String(),
				)
			}

			spanParent := span
			spanKind := trace.SpanKindServer
			ctx, span = tr.Start(ctx, "IN HTTP "+r.Method+" "+r.URL.String(),
				trace.WithAttributes(semconv_legacy.NetAttributesFromHTTPRequest("tcp", r)...),
				trace.WithAttributes(semconv_legacy.HTTPClientAttributesFromHTTPRequest(r)...),
				trace.WithAttributes(semconv_legacy.HTTPServerAttributesFromHTTPRequest(instance, routePath, r)...),
				trace.WithSpanKind(spanKind),
			)
			spanLogValues := []interface{}{FieldTraceID, span.SpanContext().TraceID().String()}
			if spanParent.SpanContext().IsValid() { // Log span.id from client side
				spanLogValues = append(spanLogValues, FieldSpanID, spanParent.SpanContext().SpanID().String())
			}
			spanLogValues = append(spanLogValues, FieldSpanID, span.SpanContext().SpanID().String())
			ctx = LogWithValues(ctx, spanLogValues...)
			Log(ctx, spanLogLevel, MsgSpanStart, FieldSpanKind, spanKind.String())

			uk := attribute.Key("username") // from HTTP header
			span.AddEvent("IN req from user", trace.WithAttributes(append(append(
				semconv_legacy.HTTPServerAttributesFromHTTPRequest(instance, routePath, r),
				semconv_legacy.HTTPClientAttributesFromHTTPRequest(r)...,
			),
				uk.String("testUser"),
			)...))

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)

			span.End()
		}

		return http.HandlerFunc(fn)
	}
}

func HttpServerLoggerMiddleware(beginLevel int, endLevel int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			routePath := getRoutePath(r)
			lrw := NewLoggingResponseWriter(w)

			Log(ctx, beginLevel, MsgInReq)
			beginTS := time.Now()

			r = r.WithContext(ctx)
			next.ServeHTTP(lrw, r)

			elapsedSec := time.Since(beginTS).Seconds()
			Log(ctx, endLevel, MsgOutResp,
				FieldUrlPattern, routePath,
				FieldStatusCode, lrw.statusCode,
				FieldReqLen, r.ContentLength,
				FieldRespLen, w.Header().Get("Content-Length"),
				FieldDuration, fmt.Sprintf("%.3f", elapsedSec),
			)
		}

		return http.HandlerFunc(fn)
	}
}

func HttpServerMetricMiddleware(ctx context.Context, meter metric_api.Meter, name string,
	description string, attributes []attribute.KeyValue,
) func(next http.Handler) http.Handler {
	baseAttrs := attributes
	attempted, err := Int64CounterGetInstrument(
		name,
		metric_api.WithDescription(description),
	)
	if err != nil {
		LogError(ctx, err, "unable to instantiate counter",
			FieldMetricName, name)
		panic(err)
	}
	durationSum, err := Float64CounterGetInstrument(
		name+"_duration",
		metric_api.WithDescription(description+", duration sum"),
		metric_api.WithUnit("s"),
	)
	if err != nil {
		LogError(ctx, err, "unable to instantiate time counter",
			FieldMetricName, name)
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			routePath := getRoutePath(r)
			lrw := NewLoggingResponseWriter(w)

			beginTS := time.Now()

			next.ServeHTTP(lrw, r)

			elapsedSec := time.Since(beginTS).Seconds()
			// To be thread-safe
			attrs := make([]attribute.KeyValue, len(baseAttrs), len(baseAttrs)+6)
			copy(attrs, baseAttrs)
			host := GetHost(r)
			attrs = append(attrs,
				attribute.Key(MetrAttrMethod).String(r.Method),
				attribute.Key(MetrAttrUrl).String(r.URL.String()),
				attribute.Key(MetrAttrHost).String(host),
				attribute.Key(MetrAttrPath).String(r.URL.Path),
				attribute.Key(MetrAttrPathPattern).String(routePath),
				attribute.Key(MetrAttrStatus).Int(lrw.statusCode),
			)
			opt := metric_api.WithAttributes(attrs...)
			attempted.Add(ctx, 1, opt)
			durationSum.Add(ctx, elapsedSec, opt)
		}

		return http.HandlerFunc(fn)
	}
}

func getRoutePath(r *http.Request) string {
	if route := mux.CurrentRoute(r); route != nil {
		if path, err := route.GetPathTemplate(); err == nil {
			return path
		}
	}
	if r.URL.RawPath != "" {
		return r.URL.RawPath
	}

	return r.URL.Path
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

type RecoveryLogger struct {
	log *slog.Logger
}

func NewRecoveryLogger(log *slog.Logger) *RecoveryLogger {
	return &RecoveryLogger{log: log}
}

func (rl *RecoveryLogger) Println(v ...any) {
	rl.log.Warn("PANIC recovered:", "error", fmt.Sprintln(v...))
}

/*
 Client HTTP middleware
*/

// LogTransport implements the http.RoundTripper interface and wraps
// outbound HTTP(S) requests with logs.
type LogTransport struct {
	rt http.RoundTripper

	beginLevel int
	endLevel   int
}

// NewLogTransport wraps the provided http.RoundTripper with one that
// logs request and respnse.
//
// If the provided http.RoundTripper is nil, http.DefaultTransport will be used
// as the base http.RoundTripper.
func NewLogTransport(base http.RoundTripper, beginLevel int, endLevel int) *LogTransport {
	if base == nil {
		base = NewHttpTransport()
	}

	return &LogTransport{
		rt:         base,
		beginLevel: beginLevel,
		endLevel:   endLevel,
	}
}

// RoundTrip logs outgoing request and response.
func (t *LogTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	var urlFull, urlHost, urlPath string
	if r.URL != nil {
		urlFull = r.URL.String()
		urlHost = r.URL.Host
		urlPath = r.URL.Path
	}
	ctx = LogWithValues(ctx,
		semconv.HTTPRequestMethodKey, r.Method,
		semconv.URLFullKey, urlFull,
		semconv.URLDomainKey, urlHost,
		semconv.URLPathKey, urlPath,
		// Skip: most probable traceID is already added to the log fields
		// FieldTraceID, span.SpanContext().TraceID(),
		FieldSpanID, span.SpanContext().SpanID(),
	)
	var res *http.Response
	var err error

	Log(ctx, t.beginLevel, MsgOutReq)

	beginTS := time.Now()

	r = r.WithContext(ctx)
	res, err = t.rt.RoundTrip(r)

	elapsedSec := time.Since(beginTS).Seconds()
	var statusCode int
	var contentLength int64
	if res != nil {
		statusCode = res.StatusCode
		contentLength = res.ContentLength
	}
	args := []any{
		FieldStatusCode, statusCode,
		FieldReqLen, r.ContentLength,
		FieldRespLen, contentLength,
		FieldDuration, fmt.Sprintf("%.3f", elapsedSec),
	}
	if err != nil {
		args = append(args, FieldError, err)
	}
	Log(ctx, t.endLevel, MsgOutResp, args...)

	return res, err //nolint:wrapcheck // should not be changed
}

// MetricTransport implements the http.RoundTripper interface and wraps
// outbound HTTP(S) requests with metrics.
type MetricTransport struct {
	rt http.RoundTripper

	meter        metric_api.Meter
	name         string
	description  string
	baseAttrs    []attribute.KeyValue
	errFormatter ErrFormatter
}

// NewMetricTransport wraps the provided http.RoundTripper with one that
// meters metrics.
//
// If the provided http.RoundTripper is nil, http.DefaultTransport will be used
// as the base http.RoundTripper.
func NewMetricTransport(base http.RoundTripper, meter metric_api.Meter, name string,
	description string, attributes []attribute.KeyValue, errFormatter ErrFormatter,
) *MetricTransport {
	if base == nil {
		base = NewHttpTransport()
	}

	return &MetricTransport{
		rt:           base,
		meter:        meter,
		name:         name,
		description:  description,
		baseAttrs:    attributes,
		errFormatter: errFormatter,
	}
}

// RoundTrip meters outgoing request-response pair.
func (t *MetricTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	attempted, err := Int64CounterGetInstrument(t.name, metric_api.WithDescription(t.description))
	if err != nil {
		LogError(ctx, err, "unable to instantiate counter", FieldMetricName, t.name)
		panic(err)
	}
	durationSum, err := Float64CounterGetInstrument(t.name+"_duration", metric_api.WithDescription(t.description+", duration sum"), metric_api.WithUnit("s"))
	if err != nil {
		LogError(ctx, err, "unable to instantiate time counter", FieldMetricName, t.name)
		panic(err)
	}
	beginTS := time.Now()
	var res *http.Response

	r = r.WithContext(ctx)
	res, err = t.rt.RoundTrip(r)

	elapsedSec := time.Since(beginTS).Seconds()
	// To be thread-safe
	attrs := make([]attribute.KeyValue, len(t.baseAttrs), len(t.baseAttrs)+6) //nolint:mnd //see append below
	copy(attrs, t.baseAttrs)
	var statusCode int
	if res != nil {
		statusCode = res.StatusCode
	}
	host := GetHost(r)
	attrs = append(attrs,
		MetrAttrMethod.String(r.Method),
		MetrAttrUrl.String(r.URL.String()),
		MetrAttrHost.String(host),
		MetrAttrPath.String(r.URL.Path),
		MetrAttrStatus.Int(statusCode),
		MetrAttrErr.String(t.errFormatter(err)),
	)
	opt := metric_api.WithAttributes(attrs...)
	attempted.Add(ctx, 1, opt)
	durationSum.Add(ctx, elapsedSec, opt)

	return res, err //nolint:wrapcheck // should not be changed
}

/*
func NewHttpClient(hostname string, serviceName string, targetServiceName string,
	appName string, captureConfig CaptureConfiger,
	log logr.Logger, logReqLevel int, logRespLevel int,
) *http.Client {
	httpClient := &http.Client{}
	return DecorateHttpClient(httpClient,
		// Trace
		map[string]string{
			SpanKeyComponent: appName,
			SpanKeyService:   serviceName,
			SpanKeyInstance:  hostname,
		},
		// Metrics
		MetrHttpOut, MetrHttpOutDescr,
		map[string]string{
			MetrAttrService:       serviceName,
			MetrAttrTargetService: targetServiceName,
		},
		// Log
		log, logReqLevel, logRespLevel,
		// Test
		captureConfig,
	)
}

func DecorateHttpClient(httpClient *http.Client,
	traceAttributes map[string]string,
	metricName string, metricDescription string, metricLabels map[string]string,
	log logr.Logger, logReqLevel int, logRespLevel int,
	captureConfig CaptureConfiger,
) *http.Client {
	attributes := []attribute.KeyValue{}
	for aKey, aVal := range traceAttributes {
		attributes = append(attributes, attribute.String(aKey, aVal))
	}
	httpClient.Transport = otelhttp.NewTransport(
		NewMetricTransport(
			NewLogTransport(
				// TODO capture
				// NewCaptureTransport(
				httpClient.Transport,

				// 	captureConfig.GetCaptureTransportMode(),
				// 	captureConfig.GetCaptureDir(),
				// 	captureConfig.GetCaptureMatchers(),
				// ),
				logReqLevel,
				logRespLevel,
			),
			GetMeter(log),
			metricName, metricDescription, metricLabels,
			FirstErr,
		),
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
		otelhttp.WithSpanOptions(trace.WithAttributes(attributes...)),
	)

	return httpClient
}
*/

func DecorateHttpTransport(ctx context.Context, httpTransport http.RoundTripper,
	traceAttributes []attribute.KeyValue,
	metricName string, metricDescription string, metricLabels []attribute.KeyValue,
	log logr.Logger, logReqLevel int, logRespLevel int,
	captureConfig CaptureConfiger,
) http.RoundTripper {
	spanService := ""
	for _, attr := range traceAttributes {
		if attr.Key == SpanKeyService {
			spanService = attr.Value.AsString()
			break
		}
	}
	return otelhttp.NewTransport(
		NewMetricTransport(
			NewLogTransport(
				// TODO capture
				// NewCaptureTransport(
				httpTransport,

				// 	captureConfig.GetCaptureTransportMode(),
				// 	captureConfig.GetCaptureDir(),
				// 	captureConfig.GetCaptureMatchers(),
				// ),
				logReqLevel,
				logRespLevel,
			),
			GetMeter(ctx),
			metricName, metricDescription, metricLabels,
			FirstErrPart,
		),
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
		otelhttp.WithSpanOptions(trace.WithAttributes(traceAttributes...)),
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			prefix := ""
			if spanService != "" {
				prefix = spanService + ": "
			}
			path := ""
			if r.URL != nil {
				path = r.URL.Path
			}
			return fmt.Sprintf("%sHTTP %s %s", prefix, r.Method, path)
		}),
	)
}

func NewHttpTransport() *http.Transport {
	// Modified from https://go.dev/src/net/http/transport.go:DefaultTransport
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ForceAttemptHTTP2: true,
		// ForceAttemptHTTP2:     false, // for more easy debugging
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
