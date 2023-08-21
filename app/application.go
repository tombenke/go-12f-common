package app

import (
	"flag"
)

// Generic Application life-cycle management functions
type LifecycleManager interface {
	GetConfigFlagSet(fs *flag.FlagSet)
	Startup()
	Shutdown()
	Check() error
}

// Telemetry related functions
type TelemetryProvider interface {
}
