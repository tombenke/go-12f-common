package apprun_test

import (
	"flag"
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
	log.Logger.Infof("TestApp GetConfigFlagSet")
}

func (a *TestApp) Startup(wg *sync.WaitGroup) error {
	log.Logger.Infof("TestApp Startup")
	a.wg = wg
	return nil
}

func (a *TestApp) Shutdown() error {
	log.Logger.Infof("TestApp Shutdown")
	return nil
}

func (a *TestApp) Check() error {
	log.Logger.Infof("TestApp Check")
	return a.err
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
		log.Logger.Infof("Start the app runner in blocking mode, that will be killed after 200 msec")
		appRunner.Run()
		twg.Done()
	}()

	twg.Add(1)
	go func() {
		log.Logger.Infof("Wait for 200 msec, then send TERM signal to the application")
		<-time.After(200 * time.Millisecond)
		// Sent TERM signal
		must.Must(syscall.Kill(syscall.Getpid(), syscall.SIGTERM))
		twg.Done()
	}()

	log.Logger.Infof("Wait for the threads to finish")
	twg.Wait()
}
