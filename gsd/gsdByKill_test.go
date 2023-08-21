package gsd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"syscall"
	"testing"
)

func TestRegisterByKill(t *testing.T) {
	var mu sync.Mutex
	gsdCbCalled := false

	wg := sync.WaitGroup{}

	// Register the callback handler
	RegisterGsdCallback(&wg, func(s os.Signal) {
		mu.Lock()
		gsdCbCalled = true
		mu.Unlock()
	})

	// Sent TERM signal to the process, then wait for termination
	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	assert.Nil(t, err)
	wg.Wait()

	// Checks if callback was called
	mu.Lock()
	assert.True(t, gsdCbCalled)
	mu.Unlock()
}
