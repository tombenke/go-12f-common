package log

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

type Logger = slog.Logger
type loggerKeyType int

const (
	// TODO change to:
	// type loggerKey struct{}
	loggerKey loggerKeyType = iota

	FieldError = "error"
)

// Setup the default logger with the given level and format
func SetupDefault(logLevel string, logFormat string) {
	leveler := slog.LevelInfo
	switch strings.ToLower(logLevel) {
	case "panic", "fatal", "error":
		leveler = slog.LevelError
	case "info":
		leveler = slog.LevelInfo
	case "warning":
		leveler = slog.LevelWarn
	case "debug", "trace":
		leveler = slog.LevelDebug
	}
	slogHandlerOptions := &slog.HandlerOptions{
		Level: leveler,
	}
	var slogHandler slog.Handler
	switch strings.ToLower(logFormat) {
	case "json":
		slogHandler = slog.NewJSONHandler(os.Stderr, slogHandlerOptions)
	case "text":
		slogHandler = slog.NewTextHandler(os.Stderr, slogHandlerOptions)
	}
	slog.SetDefault(slog.New(slogHandler))
}

// Adds fields to the logger in the context or the default one, then returns the context with the child logger
func With(ctx context.Context, args ...any) (context.Context, *slog.Logger) {
	return WithLogger(ctx, GetFromContextOrDefault(ctx), args...)
}

// Adds fields to the logger, then returns the context with the child logger
func WithLogger(ctx context.Context, logger *slog.Logger, args ...any) (context.Context, *slog.Logger) {
	newLogger := logger.With(args...)
	return context.WithValue(ctx, loggerKey, newLogger), newLogger
}

// Returns the logger stored in the context or the default one
func GetFromContextOrDefault(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// GetLogger returns the logger stored in the context or nil.
// Use FromContext to get or create a logger.
func GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return nil
}

// FromContext returns a Logger from ctx or creates it if no Logger is found.
// If it creates or there are fields, the returned context is a new child.
//
// Full example usage (logger and context will be changed, context will be passed towards):
//
//	var log logger.Logger
//	ctx, log = logger.FromContext(ctx,
//		LogKeyOutServerUri, url,
//	)
//
// Simple example usage (logger and context won't be changed):
//
//	_, log := logger.FromContext(ctx)
//
// Advanced example usage (logger and context will be changed, context won't be passed towards):
//
//	_, log := logger.LoggerFromCtx(ctx,
//		LogKeyOutServerUri, url,
//	)
func FromContext(ctx context.Context, keysAndValues ...any) (context.Context, *slog.Logger) {
	if ctx == nil {
		ctx = context.Background()
	}
	var log *slog.Logger
	var has bool
	var store bool
	if log, has = ctx.Value(loggerKey).(*slog.Logger); !has || log == nil {
		log = GetFromContextOrDefault(ctx)
		store = true
	}
	if len(keysAndValues) > 0 {
		log = log.With(keysAndValues...)
		store = true
	}
	if store {
		ctx = NewContext(ctx, log)
	}

	return ctx, log
}

func FromContextKeyValue(ctx context.Context, keysAndValues ...attribute.KeyValue) (context.Context, *slog.Logger) {
	values := make([]any, 0, len(keysAndValues)*2)
	for _, kv := range keysAndValues {
		values = append(values, kv.Key, kv.Value.Emit())
	}

	return FromContext(ctx, values...)
}

// NewContext returns a new Context, derived from ctx, which carries the
// provided Logger.
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		logger = GetFromContextOrDefault(ctx)
	}

	return context.WithValue(ctx, loggerKey, logger)
}

// Logs with the logger in the context or the default one at info level
func InfoContext(ctx context.Context, msg string, args ...any) {
	logger := GetFromContextOrDefault(ctx)
	logger.InfoContext(ctx, msg, args...)
}

// Logs with the logger in the context or the default one at warn level
func WarnContext(ctx context.Context, msg string, args ...any) {
	logger := GetFromContextOrDefault(ctx)
	logger.WarnContext(ctx, msg, args...)
}

// Logs with the logger in the context or the default one at debug level
func DebugContext(ctx context.Context, msg string, args ...any) {
	logger := GetFromContextOrDefault(ctx)
	logger.DebugContext(ctx, msg, args...)
}

// Logs with the logger in the context or the default one at error level
func ErrorContext(ctx context.Context, msg string, args ...any) {
	logger := GetFromContextOrDefault(ctx)
	logger.ErrorContext(ctx, msg, args...)
}
