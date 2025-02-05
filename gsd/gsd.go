// Package gsd is a simple package to manage graceful shut-down via catching the terminations signals
package gsd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/tombenke/go-12f-common/log"
)

// Register is an observer go routine to get notifed when termination signals arrive,
// then call the `cb` callback function with the signal and finishes the go routine.
func RegisterGsdCallback(ctx context.Context, wg *sync.WaitGroup, cb func(os.Signal)) chan os.Signal {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		// Block until a signal is received.
		for {
			s := <-sigs
			log.DebugContext(ctx, "Got signal", "signal", s)
			if s == syscall.SIGINT || s == syscall.SIGTERM {
				wg.Done()
				cb(s)
				break
			}
		}
	}()

	return sigs
}
