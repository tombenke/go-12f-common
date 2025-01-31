package slog

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

func AppendToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

func Default() *slog.Logger {
	return slog.Default()
}

func GetFromContextOrDefault(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	logger := GetFromContextOrDefault(ctx)
	logger.InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	logger := GetFromContextOrDefault(ctx)
	logger.WarnContext(ctx, msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	logger := GetFromContextOrDefault(ctx)
	logger.DebugContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	logger := GetFromContextOrDefault(ctx)
	logger.ErrorContext(ctx, msg, args...)
}
