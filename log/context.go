package log

import (
	"context"
	"errors"
	"sync"

	"github.com/sirupsen/logrus"
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
func FromContext(ctx context.Context, keysAndValues logrus.Fields) (context.Context, FieldLogger) {
	if ctx == nil {
		ctx = context.Background()
	}
	var logger FieldLogger
	var has bool
	var store bool
	if logger, has = ctx.Value(contextKey{}).(FieldLogger); !has || logger == nil {
		logger = GetLogger(unknownAppName, logrus.DebugLevel)
		store = true
	}
	if len(keysAndValues) > 0 {
		logger = logger.WithFields(keysAndValues)
		store = true
	}
	if store {
		ctx = NewContext(ctx, logger)
	}

	return ctx, logger
}

// NewContext returns a new Context, derived from ctx, which carries the
// provided Logger.
func NewContext(ctx context.Context, logger logrus.FieldLogger) context.Context {
	if logger == nil {
		logger = GetLogger(unknownAppName, logrus.DebugLevel)
	}

	return context.WithValue(ctx, contextKey{}, logger)
}

// GetLogger returns a registered logger with app name.
// Creates a new instance, if not exists (uses the level only in this case)
func GetLogger(app string, level logrus.Level) FieldLogger {
	if logger, has := loggers.Load(app); has {
		return logger.(FieldLogger) //nolint:forcetypeassert // always *slog.Logger
	}
	newFormatter := BuildFormatter("text")
	//newFormatter := GetFormatter("json")

	logger := &LogrusLogger{logrus.New()}
	logger.SetFormatter(newFormatter)
	logger.SetLevel(level)
	logger.WithField("logger", app)
	loggers.Store(app, logger)

	return logger
}
