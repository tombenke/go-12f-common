package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type Logger = slog.Logger
type LoggerKeyType int

const (
	LoggerKey LoggerKeyType = iota
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
	return context.WithValue(ctx, LoggerKey, newLogger), newLogger
}

// Returns the logger stored in the context or the default one
func GetFromContextOrDefault(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
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
