package oti

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/tombenke/go-12f-common/v2/buildinfo"
	"github.com/tombenke/go-12f-common/v2/log"
)

// NatsCarrier adapts nats.Header to satisfy the TextMapCarrier interface.
type NatsCarrier nats.Header

// Get returns the value associated with the passed key.
func (hc NatsCarrier) Get(key string) string {
	return nats.Header(hc).Get(key)
}

// Set stores the key-value pair.
func (hc NatsCarrier) Set(key string, value string) {
	nats.Header(hc).Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (hc NatsCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range hc {
		keys = append(keys, k)
	}
	return keys
}

// ObsRequestMsg observes the NATS RequestMsg calls
func ObsRequestMsg(ctx context.Context, fn func(msg *nats.Msg, timeout time.Duration) (*nats.Msg, error),
	serviceName string, netLayer string,
	startLevel int, endLevel int,
) func(msg *nats.Msg, timeout time.Duration) (*nats.Msg, error) {
	return func(msg *nats.Msg, timeout time.Duration) (*nats.Msg, error) {
		meter := GetMeter(ctx)
		appName := buildinfo.AppName()
		appVersion := buildinfo.Version()
		SubscrSubject := ""
		SubscrQueue := ""
		if msg.Sub != nil {
			SubscrSubject = msg.Sub.Subject
			SubscrQueue = msg.Sub.Queue
		}
		spanName := fmt.Sprintf("nats.Conn.RequestMsg %s %s/%s %s:%s",
			appName, msg.Subject, msg.Reply, SubscrSubject, SubscrQueue)

		/* Belows were set to otel.GetTracerProvider() and rootCtx logger at startup

		semconv.ServiceNamespaceKey.String(podNamespace),
		FieldApp.String(buildinfo.AppName()),
		semconv.ServiceInstanceIDKey.String(hostname),
		semconv.ServiceVersionKey.String(buildinfo.Version()),
		*/

		trConf := trace.WithInstrumentationAttributes(
			FieldService.String(appName+"."+serviceName),
			FieldSubject.String(msg.Subject),
			FieldSubscrSubject.String(SubscrSubject),
			FieldSubscrQueue.String(SubscrQueue),
			FieldReply.String(msg.Reply),
			FieldNetLayer.String(netLayer),
			FieldNetOut,
		)
		tr := otel.GetTracerProvider().Tracer(serviceName, trConf)
		mws := []InternalMiddleware[*nats.Msg]{
			/*TryCatch[*nats.Msg]()*/
			Span[*nats.Msg](tr, trace.SpanKindServer, spanName),
			Logger[*nats.Msg](
				"nats.Conn.RequestMsg START", startLevel,
				"nats.Conn.RequestMsg END", endLevel,
				"nats.Conn.RequestMsg ERROR",
				[]attribute.KeyValue{
					FieldService.String(appName + "." + serviceName),
					FieldSubject.String(msg.Subject),
					FieldSubscrSubject.String(SubscrSubject),
					FieldSubscrQueue.String(SubscrQueue),
					FieldReply.String(msg.Reply),
					FieldNetLayer.String(netLayer),
					FieldNetOut,
				}),
			Metrics[*nats.Msg](ctx, meter, "nlhc_out_nats_requestmsg", "Enqueue Requests", []attribute.KeyValue{
				FieldApp.String(appName),
				FieldService.String(serviceName),
				FieldVersion.String(appVersion),
				FieldSubject.String(msg.Subject),
				FieldSubscrSubject.String(SubscrSubject),
				FieldSubscrQueue.String(SubscrQueue),
				// FieldReply.String(msg.Reply), // may be vary (cardinality!)
				FieldNetLayer.String(netLayer),
				FieldNetOut,
			}, FirstErrPart),
			//TryCatch[*nats.Msg](),
		}

		return InternalMiddlewareChain(mws...)(func(ctxFn context.Context) (*nats.Msg, error) {
			carrier := NatsCarrier{}
			maps.Copy(NatsCarrier(msg.Header), carrier)
			otel.GetTextMapPropagator().Inject(ctxFn, carrier)
			msg.Header = nats.Header(carrier)

			return fn(msg, timeout)
		})(ctx)
	}
}

// ObsNatsMsgHandler observes the NATS MsgHandler calls
// Adds ctx and error to the original function signature:
// type MsgHandler func(msg *Msg) error
func ObsNatsMsgHandler(rootCtx context.Context, fn func(context.Context, *nats.Msg) error,
	serviceName string, netLayer string,
	startLevel int, endLevel int,
) nats.MsgHandler {
	return func(msg *nats.Msg) {
		// TODO handle shutdown rootCtx
		// TODO handle fn timeout

		ctx := context.Background()
		if msg.Header.Get(TraceparentHeader) != "" {
			ctx = otel.GetTextMapPropagator().Extract(rootCtx, NatsCarrier(msg.Header))
			// oti.Span does not set traceID spanID to the context logger, if it's already present, set it here
			spanParentCtx := trace.SpanFromContext(ctx).SpanContext()
			ctx, _ = log.FromContext(ctx,
				string(FieldTraceID), spanParentCtx.TraceID().String(),
				string(FieldSpanID), spanParentCtx.SpanID().String(),
			)
		}

		meter := GetMeter(rootCtx)
		appName := buildinfo.AppName()
		appVersion := buildinfo.Version()
		SubscrSubject := ""
		SubscrQueue := ""
		if msg.Sub != nil {
			SubscrSubject = msg.Sub.Subject
			SubscrQueue = msg.Sub.Queue
		}
		spanName := fmt.Sprintf("nats.Conn.RequestMsg %s %s/%s %s:%s",
			appName, msg.Subject, msg.Reply, SubscrSubject, SubscrQueue)

		/* Belows were set to otel.GetTracerProvider() and rootCtx logger at startup

		semconv.ServiceNamespaceKey.String(podNamespace),
		FieldApp.String(buildinfo.AppName()),
		semconv.ServiceInstanceIDKey.String(hostname),
		semconv.ServiceVersionKey.String(buildinfo.Version()),
		*/

		trConf := trace.WithInstrumentationAttributes(
			FieldService.String(appName+"."+serviceName),
			FieldSubject.String(msg.Subject),
			FieldSubscrSubject.String(SubscrSubject),
			FieldSubscrQueue.String(SubscrQueue),
			FieldReply.String(msg.Reply),
			FieldNetLayer.String(netLayer),
			FieldNetIn,
		)
		tr := otel.GetTracerProvider().Tracer(serviceName, trConf)
		mws := []InternalMiddleware[any]{
			/*TryCatch[*nats.Msg]()*/
			Span[any](tr, trace.SpanKindServer, spanName),
			Logger[any](
				"nats.Conn.Subscribe.Receive START", startLevel,
				"nats.Conn.Subscribe.Receive END", endLevel,
				"nats.Conn.Subscribe.Receive ERROR",
				[]attribute.KeyValue{
					FieldService.String(appName + "." + serviceName),
					FieldSubject.String(msg.Subject),
					FieldSubscrSubject.String(SubscrSubject),
					FieldSubscrQueue.String(SubscrQueue),
					FieldReply.String(msg.Reply),
					FieldNetLayer.String(netLayer),
					FieldNetIn,
				}),
			Metrics[any](rootCtx, meter, "nlhc_in_nats_handlemsg", "Handle Message", []attribute.KeyValue{
				FieldApp.String(appName),
				FieldService.String(appName + "." + serviceName),
				FieldVersion.String(appVersion),
				FieldSubject.String(msg.Subject),
				FieldSubscrSubject.String(SubscrSubject),
				FieldSubscrQueue.String(SubscrQueue),
				// FieldReply.String(msg.Reply), // may be vary (cardinality!)
				FieldNetLayer.String(netLayer),
				FieldNetIn,
			}, FirstErrPart),
			//TryCatch[*nats.Msg](),
		}

		_, _ = InternalMiddlewareChain(mws...)(func(ctxFn context.Context) (any, error) {

			err := fn(ctxFn, msg)

			return nil, err
		})(ctx)
	}
}

// ObsPublishMsg observes the NATS JS PublishMsg calls
// (ctx context.Context, msg *nats.Msg, opts ...PublishOpt) (*PubAck, error)
func ObsPublishMsg(ctx context.Context, fn func(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error),
	serviceName string, netLayer string,
	startLevel int, endLevel int,
) func(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
	return func(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
		meter := GetMeter(ctx)
		appName := buildinfo.AppName()
		appVersion := buildinfo.Version()
		SubscrSubject := ""
		SubscrQueue := ""
		if msg.Sub != nil {
			SubscrSubject = msg.Sub.Subject
			SubscrQueue = msg.Sub.Queue
		}
		spanName := fmt.Sprintf("nats.jetstream.Publisher.PublishMsg %s %s/%s %s:%s",
			appName, msg.Subject, msg.Reply, SubscrSubject, SubscrQueue)
		msgType := msg.Header.Get(FieldMessageType)

		/* Belows were set to otel.GetTracerProvider() and rootCtx logger at startup

		semconv.ServiceNamespaceKey.String(podNamespace),
		FieldApp.String(buildinfo.AppName()),
		semconv.ServiceInstanceIDKey.String(hostname),
		semconv.ServiceVersionKey.String(buildinfo.Version()),
		*/

		trConf := trace.WithInstrumentationAttributes(
			FieldService.String(appName+"."+serviceName),
			FieldSubject.String(msg.Subject),
			FieldSubscrSubject.String(SubscrSubject),
			FieldSubscrQueue.String(SubscrQueue),
			FieldReply.String(msg.Reply),
			FieldNetLayer.String(netLayer),
			FieldNetOut,
			FieldMsgType.String(msgType),
		)
		tr := otel.GetTracerProvider().Tracer(serviceName, trConf)
		mws := []InternalMiddleware[*jetstream.PubAck]{
			/*TryCatch[*nats.Msg]()*/
			Span[*jetstream.PubAck](tr, trace.SpanKindServer, spanName),
			Logger[*jetstream.PubAck](
				"nats.jetstream.Publisher.PublishMsg START", startLevel,
				"nats.jetstream.Publisher.PublishMsg END", endLevel,
				"nats.jetstream.Publisher.PublishMsg ERROR",
				[]attribute.KeyValue{
					FieldService.String(appName + "." + serviceName),
					FieldSubject.String(msg.Subject),
					FieldSubscrSubject.String(SubscrSubject),
					FieldSubscrQueue.String(SubscrQueue),
					FieldReply.String(msg.Reply),
					FieldNetLayer.String(netLayer),
					FieldNetOut,
					FieldMsgType.String(msgType),
				}),
			Metrics[*jetstream.PubAck](ctx, meter, "nlhc_out_nats_js_publishmsg", "Publish Msg", []attribute.KeyValue{
				FieldApp.String(appName),
				FieldService.String(serviceName),
				FieldVersion.String(appVersion),
				FieldSubject.String(msg.Subject),
				FieldSubscrSubject.String(SubscrSubject),
				FieldSubscrQueue.String(SubscrQueue),
				// FieldReply.String(msg.Reply), // may be vary (cardinality!)
				FieldNetLayer.String(netLayer),
				FieldNetOut,
				FieldMsgType.String(msgType),
			}, FirstErrPart),
			//TryCatch[*nats.Msg](),
		}

		return InternalMiddlewareChain(mws...)(func(ctxFn context.Context) (*jetstream.PubAck, error) {
			carrier := NatsCarrier{}
			maps.Copy(NatsCarrier(msg.Header), carrier)
			otel.GetTextMapPropagator().Inject(ctxFn, carrier)
			msg.Header = nats.Header(carrier)

			return fn(ctxFn, msg, opts...)
		})(ctx)
	}
}

func ObsNatsConsumerMessageHandler(rootCtx context.Context, fn func(context.Context, jetstream.Msg) error,
	serviceName string,
	startLevel int, endLevel int,
) func(jetstream.Msg) {
	return func(msg jetstream.Msg) {
		// TODO handle shutdown rootCtx
		// TODO handle fn timeout

		ctx := context.Background()
		if msg.Headers().Get(TraceparentHeader) != "" {
			ctx = otel.GetTextMapPropagator().Extract(rootCtx, NatsCarrier(msg.Headers()))
			// oti.Span does not set traceID spanID to the context logger, if it's already present, set it here
			spanParentCtx := trace.SpanFromContext(ctx).SpanContext()
			ctx, _ = log.FromContext(ctx,
				string(FieldTraceID), spanParentCtx.TraceID().String(),
				string(FieldSpanID), spanParentCtx.SpanID().String(),
			)
		}

		meter := GetMeter(rootCtx)
		appName := buildinfo.AppName()
		appVersion := buildinfo.Version()
		spanName := fmt.Sprintf("nats.Conn.RequestMsg %s %s/%s",
			appName, msg.Subject(), msg.Reply())

		/* Belows were set to otel.GetTracerProvider() and rootCtx logger at startup

		semconv.ServiceNamespaceKey.String(podNamespace),
		FieldApp.String(buildinfo.AppName()),
		semconv.ServiceInstanceIDKey.String(hostname),
		semconv.ServiceVersionKey.String(buildinfo.Version()),
		*/

		trConf := trace.WithInstrumentationAttributes(
			FieldService.String(appName+"."+serviceName),
			FieldSubject.String(msg.Subject()),
			FieldReply.String(msg.Reply()),
			FieldNetIn,
		)
		tr := otel.GetTracerProvider().Tracer(serviceName, trConf)
		mws := []InternalMiddleware[any]{
			/*TryCatch[*nats.Msg]()*/
			Span[any](tr, trace.SpanKindServer, spanName),
			Logger[any](
				"nats.Conn.Subscribe.Receive START", startLevel,
				"nats.Conn.Subscribe.Receive END", endLevel,
				"nats.Conn.Subscribe.Receive ERROR",
				[]attribute.KeyValue{
					FieldService.String(appName + "." + serviceName),
					FieldSubject.String(msg.Subject()),
					FieldReply.String(msg.Reply()),
					FieldNetIn,
				}),
			Metrics[any](rootCtx, meter, "nlhc_in_nats_js_consumemsg", "Consume Message", []attribute.KeyValue{
				FieldApp.String(appName),
				FieldService.String(appName + "." + serviceName),
				FieldVersion.String(appVersion),
				FieldSubject.String(msg.Subject()),
				// FieldReply.String(msg.Reply), // may be vary (cardinality!)
				FieldNetIn,
			}, FirstErrPart),
			//TryCatch[*nats.Msg](),
		}

		_, _ = InternalMiddlewareChain(mws...)(func(ctxFn context.Context) (any, error) {

			err := fn(ctxFn, msg)

			return nil, err
		})(ctx)
	}
}
