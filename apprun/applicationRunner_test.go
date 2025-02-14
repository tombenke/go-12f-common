package apprun_test

import (
	"context"
	"log/slog"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"github.com/tombenke/go-12f-common/v2/apprun"
	"github.com/tombenke/go-12f-common/v2/must"
)

type Config struct{}

type TestApp struct {
}

func NewTestApp() apprun.Application {
	return &TestApp{}
}

func (a *TestApp) Components(ctx context.Context) []apprun.ComponentLifecycleManager {
	return nil
}

func TestApplicationRunner_StartStop(t *testing.T) {
	testApp := NewTestApp()

	flagSet := pflag.NewFlagSet("root", pflag.ContinueOnError)
	config := &apprun.Config{}
	config.GetConfigFlagSet(flagSet)
	require.NoError(t, config.LoadConfig(flagSet))
	appRunner := apprun.NewApplicationRunner(config, testApp)

	twg := &sync.WaitGroup{}

	twg.Add(1)
	go func() {
		slog.Info("Start the app runner in blocking mode, that will be killed after 200 msec")
		require.NoError(t, appRunner.Run())
		twg.Done()
	}()

	twg.Add(1)
	go func() {
		slog.Info("Wait for 200 msec, then send TERM signal to the application")
		<-time.After(200 * time.Millisecond)
		// Sent TERM signal
		must.Must(syscall.Kill(syscall.Getpid(), syscall.SIGTERM))
		twg.Done()
	}()

	slog.Info("Wait for the threads to finish")
	twg.Wait()
}
