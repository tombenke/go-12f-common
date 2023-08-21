package app

import (
	"flag"

	"github.com/tombenke/go-12f-common/env"
)

const (
	ShowHelpHelp    = "Show help message"
	ShowHelpDefault = false

	PrintConfigHelp    = "Print configuration parameters"
	PrintConfigEnvVar  = "PRINT_CONFIG"
	PrintConfigDefault = false

	logLevelHelp    = "The log level: panic | fatal | error | warning | info | debug | trace"
	logLevelEnvVar  = "LOG_LEVEL"
	defaultLogLevel = "info"

	logFormatHelp    = "The log format: json | text"
	logFormatEnvVar  = "LOG_FORMAT"
	defaultLogFormat = "json"

	HealthCheckPortHelp    = "The HTTP port of the healthcheck endpoints"
	HealthCheckPortEnvVar  = "HEALTHCHECK_PORT"
	HealthCheckPortDefault = "8080"

	LivenessCheckPathHelp    = "The path of the liveness check endpoint"
	LivenessCheckPathEnvVar  = "LIVENESS_CHECK_PATH"
	LivenessCheckPathDefault = "/live"

	ReadinessCheckPathHelp    = "The path of the readiness check endpoint"
	ReadinessCheckPathEnvVar  = "READINESS_CHECK_PATH"
	ReadinessCheckPathDefault = "/ready"
)

type Config struct {
	LogLevel           string
	LogFormat          string
	ShowHelp           bool
	PrintConfig        bool
	HealthCheckPort    int
	LivenessCheckPath  string
	ReadinessCheckPath string
}

func (cfg *Config) GetConfigFlagSet(fs *flag.FlagSet) {

	// Generic config parameters and CLI flags
	fs.BoolVar(&cfg.ShowHelp, "h", ShowHelpDefault, ShowHelpHelp)
	fs.BoolVar(&cfg.ShowHelp, "help", ShowHelpDefault, ShowHelpHelp)

	fs.BoolVar(&cfg.PrintConfig, "p", PrintConfigDefault, PrintConfigHelp)
	fs.BoolVar(&cfg.PrintConfig, "print-config", PrintConfigDefault, PrintConfigHelp)

	// Logger parameters
	fs.StringVar(&cfg.LogLevel, "l", env.GetEnvWithDefault(logLevelEnvVar, defaultLogLevel), logLevelHelp)
	fs.StringVar(&cfg.LogLevel, "log-level", env.GetEnvWithDefault(logLevelEnvVar, defaultLogLevel), logLevelHelp)

	fs.StringVar(&cfg.LogFormat, "f", env.GetEnvWithDefault(logFormatEnvVar, defaultLogFormat), logFormatHelp)
	fs.StringVar(&cfg.LogFormat, "log-format", env.GetEnvWithDefault(logFormatEnvVar, defaultLogFormat), logFormatHelp)

	// HealthCheck parameters
	fs.IntVar(&cfg.HealthCheckPort, "healthcheck-port", int(env.GetEnvWithDefaultUint(HealthCheckPortEnvVar, HealthCheckPortDefault)), HealthCheckPortHelp)
	fs.StringVar(&cfg.LivenessCheckPath, "liveness-check-path", env.GetEnvWithDefault(LivenessCheckPathEnvVar, LivenessCheckPathDefault), LivenessCheckPathHelp)
	fs.StringVar(&cfg.ReadinessCheckPath, "readiness-check-path", env.GetEnvWithDefault(ReadinessCheckPathEnvVar, ReadinessCheckPathDefault), ReadinessCheckPathHelp)

	// OTEL parameters
	// TODO
}
