package app

import (
	"flag"
	"sync"
)

// Generic Application life-cycle management functions
type LifecycleManager interface {
	GetConfigFlagSet(fs *flag.FlagSet)
	Startup(wg *sync.WaitGroup)
	Shutdown()
	Check() error
}
