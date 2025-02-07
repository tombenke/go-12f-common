package oti

// TODO:
//// Initializes an OTLP exporter, and configures the corresponding trace provider.
//func initTracerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (func(context.Context) error, error) {
//	// Set up a trace exporter
//	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
//	if err != nil {
//		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
//	}
//
//	// Register the trace exporter with a TracerProvider, using a batch
//	// span processor to aggregate spans before export.
//	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
//	tracerProvider := sdktrace.NewTracerProvider(
//		sdktrace.WithSampler(sdktrace.AlwaysSample()),
//		sdktrace.WithResource(res),
//		sdktrace.WithSpanProcessor(bsp),
//	)
//	otel.SetTracerProvider(tracerProvider)
//
//	// Set global propagator to tracecontext (the default is no-op).
//	otel.SetTextMapPropagator(propagation.TraceContext{})
//
//	// Shutdown will flush any remaining spans and shut down the exporter.
//	return tracerProvider.Shutdown, nil
//}
