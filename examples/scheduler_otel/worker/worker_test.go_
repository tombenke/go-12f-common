package worker

import (
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

// Create a Worker
func createWorker(t *testing.T) (*Worker, *sync.WaitGroup) {
	worker, err := NewWorker(
		&Config{},
		make(chan time.Time),
	)
	assert.Nil(t, err)
	fs := &pflag.FlagSet{}
	worker.GetConfigFlagSet(fs)
	wg := &sync.WaitGroup{}
	worker.Startup(wg)
	return worker, wg
}

// Create, Start and stop a worker instance
func TestWorkerStartStop(t *testing.T) {
	worker, wg := createWorker(t)
	worker.Shutdown()
	wg.Wait()
}
