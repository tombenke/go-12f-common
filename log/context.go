package log

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
)

const (
	KeyError       = "error"
	KeyCmd         = "command"
	KeyTestCase    = "testcase"
	KeyApp         = "app"
	KeyService     = "service"
	unknownAppName = "unknown"
)

var ErrInvalidConfig = errors.New("invalid config")

var loggers = sync.Map{} //nolint:gochecknoglobals // simple logging

// contextKey is how we find Loggers in a context.Context.
type contextKey struct{}

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
func FromContext(ctx context.Context, keysAndValues ...interface{}) (context.Context, *slog.Logger) {
	if ctx == nil {
		ctx = context.Background()
	}
	var log *slog.Logger
	var has bool
	var store bool
	if log, has = ctx.Value(contextKey{}).(*slog.Logger); !has || log == nil {
		log = GetLogger(unknownAppName, slog.LevelDebug)
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

// NewContext returns a new Context, derived from ctx, which carries the
// provided Logger.
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		logger = GetLogger(unknownAppName, slog.LevelDebug)
	}

	return context.WithValue(ctx, contextKey{}, logger)
}

// GetLogger returns a registered logger with app name.
// Creates a new instance, if not exists (uses the level only in this case)
func GetLogger(app string, level slog.Level) *slog.Logger {
	if logger, has := loggers.Load(app); has {
		return logger.(*slog.Logger) //nolint:forcetypeassert // always *slog.Logger
	}
	newHandler := slog.NewTextHandler
	//newHandler := slog.NewJSONHandler

	logger := slog.New(newHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})).With("logger", app)
	loggers.Store(app, logger)

	return logger
}
