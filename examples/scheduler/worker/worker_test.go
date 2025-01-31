package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

// Create a Worker
func createWorker(t *testing.T) (*Worker, *sync.WaitGroup) {
	worker := NewWorker(
		&Config{},
		make(chan time.Time),
	)
	fs := &pflag.FlagSet{}
	worker.GetConfigFlagSet(fs)
	wg := &sync.WaitGroup{}
	require.NoError(t, worker.Startup(context.Background(), wg))
	return worker, wg
}

// Create, Start and stop a worker instance
func TestWorkerStartStop(t *testing.T) {
	worker, wg := createWorker(t)
	require.NoError(t, worker.Shutdown(context.Background()))
	wg.Wait()
}
