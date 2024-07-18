package scheduler

import (
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

// Create a Scheduler
func createScheduler(t *testing.T) (*Scheduler, *sync.WaitGroup) {
	scheduler, err := NewScheduler(
		&Config{},
		make(chan time.Time),
	)
	assert.Nil(t, err)
	fs := &pflag.FlagSet{}
	scheduler.GetConfigFlagSet(fs)
	wg := &sync.WaitGroup{}
	scheduler.Startup(wg)
	return scheduler, wg
}

// Create, Start and stop a scheduler instance
func TestSchedulerStartStop(t *testing.T) {
	scheduler, wg := createScheduler(t)
	scheduler.Shutdown()
	wg.Wait()
}
