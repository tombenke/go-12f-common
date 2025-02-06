package apprun_test

import (
	"context"
	"flag"
	"log/slog"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"
)

type Config struct{}

type TestApp struct {
	wg  *sync.WaitGroup
	err error
}

func NewTestApp() apprun.LifecycleManager {
	return &TestApp{err: healthcheck.ServiceNotAvailableError{}}
}

func (a *TestApp) GetConfigFlagSet(fs *flag.FlagSet) {
	a.getLogger(context.Background()).Info("GetConfigFlagSet")
}

func (a *TestApp) Startup(ctx context.Context, wg *sync.WaitGroup) error {
	a.getLogger(ctx).Info("Startup")
	a.wg = wg
	return nil
}

func (a *TestApp) Shutdown(ctx context.Context) error {
	a.getLogger(ctx).Info("Shutdown")
	return nil
}

func (a *TestApp) Check(ctx context.Context) error {
	a.getLogger(ctx).Info("Check")
	return a.err
}

func (a *TestApp) getLogger(ctx context.Context) *slog.Logger {
	return log.GetFromContextOrDefault(ctx).With("app", "TestApp")
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
