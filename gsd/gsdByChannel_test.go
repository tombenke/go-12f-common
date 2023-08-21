package gsd

import (
	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/log"
	"os"
	"sync"
	"syscall"
	"testing"
)

func TestRegisterByChannel(t *testing.T) {
	log.SetLevelStr("debug")
	var mu sync.Mutex
	gsdCbCalled := false

	wg := sync.WaitGroup{}

	// Register the callback handler
	sigsCh := RegisterGsdCallback(&wg, func(s os.Signal) {
		mu.Lock()
		gsdCbCalled = true
		mu.Unlock()
	})

	// Send signals via the channel
	// Sent irrelevant signals that gsd should not catch
	sigsCh <- syscall.SIGIO
	sigsCh <- syscall.SIGUSR1
	sigsCh <- syscall.SIGQUIT

	// Sent TERM signal, then wait for termination
	sigsCh <- syscall.SIGTERM
	wg.Wait()

	// Checks if callback was called
	mu.Lock()
	assert.True(t, gsdCbCalled)
	mu.Unlock()
}
