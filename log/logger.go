package log

import (
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"runtime"
)

const (
	// OperationResultSuccess defines the result value for successful operations
	OperationResultSuccess OperationResult = "success"
	// OperationResultError defines the result value for operations failed with normal error
	OperationResultError OperationResult = "error"
	// OperationResultFatal defines the result value for operations failed with fatal error
	OperationResultFatal OperationResult = "fatal"
)

// OperationResult is the type of the possible results of the operations of the business process implementation
type OperationResult string

// Logger is the global logger
var (
	Logger *logrus.Logger

	TimestampFormat = time.RFC3339Nano
)

func init() {
	Logger = logrus.New()
}

// SetFormatterStr sets the log format to either `json` or `text`
func SetFormatterStr(format string) {
	switch strings.ToLower(format) {
	case "json":
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: TimestampFormat,
		})
	case "text":
	default:
		Logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: TimestampFormat,
		})
	}
}

// SetLevelStr sets the log level according to the `level` string parameter
func SetLevelStr(level string) {
	switch strings.ToLower(level) {
	case "panic":
		Logger.SetLevel(logrus.PanicLevel)
	case "fatal":
		Logger.SetLevel(logrus.FatalLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	case "warning":
		Logger.SetLevel(logrus.WarnLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "trace":
		Logger.SetLevel(logrus.TraceLevel)
	}
}

// WithBCID adds BCID and the actorName to the logger.
func WithBCID(l logrus.FieldLogger, bcid, applicationName string) logrus.FieldLogger {
	return l.WithFields(logrus.Fields{
		"bcid":  bcid,
		"actor": applicationName,
	})
}

// WithProcessingState adds state information to the logger.
func WithProcessingState(l logrus.FieldLogger, op string, result OperationResult) logrus.FieldLogger {
	return l.WithFields(logrus.Fields{
		"tag":       "process",
		"operation": op,
		"result":    string(result),
		"status":    op + "_" + string(result),
	})
}

// Get the name of the caller function
func getCaller() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	details := runtime.FuncForPC(pc)
	if details == nil {
		return ""
	}

	name := details.Name()
	_, line := details.FileLine(pc)

	return name + ":" + strconv.Itoa(line)
}
