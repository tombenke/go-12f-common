package oti

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	logger "github.com/tombenke/go-12f-common/v2/log"
	"go.opentelemetry.io/otel/attribute"
)

var (
	// ReplaceDotToUs controls whether to replace dots with underscores in log keys
	// Reason: Loki will replace dots with underscores,
	// so to keep consistency, it will be visible on console, too.
	ReplaceDotToUs = true
)

func LogWithValues(ctx context.Context, args ...any) context.Context {
	args = PatchLogArgs(args)
	logR, errR := logr.FromContext(ctx)
	if errR == nil { // logr
		return logr.NewContext(ctx, logR.WithValues(args...))
	}

	// slog-based go-12f-common
	ctx, _ = logger.FromContext(ctx, args...)

	return ctx
}

func CopyLogger(ctxTo context.Context, ctxFrom context.Context) context.Context {
	logR, errR := logr.FromContext(ctxFrom)
	if errR == nil { // logr
		return logr.NewContext(ctxTo, logR)
	}

	// slog-based go-12f-common
	_, logS := logger.FromContext(ctxFrom)
	return logger.NewContext(ctxTo, logS)
}

// PatchLogArgs ensures that all keys are strings (zap and slog require string keys)
// and replaces '.' with '_' in the keys if ReplaceDotToUs is true
func PatchLogArgs(args []any) []any {
	argsPatch := make([]any, 0, len(args))
	for i, arg := range args {
		if i%2 == 0 { // even indices are keys; coerce to string & sanitize
			if arg == nil {
				arg = DotToUs("<nil>")
			} else {
				switch k := arg.(type) {
				case string:
					arg = DotToUs(k)
				case attribute.Key:
					arg = DotToUs(string(k))
				case fmt.Stringer:
					// NOTE: not all fmt.Stringer impls guarantee panic-free String() on nil receivers
					arg = DotToUs(k.String())
				case fmt.GoStringer:
					arg = DotToUs(k.GoString())
				default:
					val := reflect.ValueOf(arg)
					if !val.IsValid() {
						arg = DotToUs("<invalid>")
					} else if val.Kind() == reflect.String {
						arg = DotToUs(val.Convert(reflect.TypeOf("")).Interface().(string))
					} else {
						arg = DotToUs(fmt.Sprintf("%T:%v", arg, arg))
					}
				}
			}
		}
		argsPatch = append(argsPatch, arg)
	}

	return argsPatch
}

func Log(ctx context.Context, level int, msg string, args ...any) {
	args = PatchLogArgs(args)
	logR, errR := logr.FromContext(ctx)
	if errR == nil { // logr
		logR.V(level).Info(msg, args...)

		return
	}

	// slog-based go-12f-common
	_, logS := logger.FromContext(ctx)
	logS.Log(ctx, slog.Level(level), msg, args...)
}

func LogError(ctx context.Context, err error, msg string, args ...any) {
	args = PatchLogArgs(args)
	logR, errR := logr.FromContext(ctx)
	if errR == nil { // logr
		logR.Error(err, msg, args...)

		return
	}

	// slog-based go-12f-common
	_, logS := logger.FromContext(ctx, args...)
	logS.Log(ctx, slog.LevelError, msg, string(FieldError), err)
}

func DotToUsIf[T ~string](s T, dotToUs bool) string {
	if dotToUs {
		return strings.ReplaceAll(string(s), ".", "_")
	}
	return string(s)
}

func DotToUs[T ~string](s T) string {
	if ReplaceDotToUs {
		return strings.ReplaceAll(string(s), ".", "_")
	}
	return string(s)
}
