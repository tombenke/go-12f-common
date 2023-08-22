package apprun_test

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"
	"sync"
	"syscall"
	"testing"
	"time"
)

type TestApp struct {
	wg  *sync.WaitGroup
	err error
}

func NewTestApp() (apprun.LifecycleManager, error) {
	return &TestApp{err: healthcheck.ServiceNotAvailableError{}}, nil
}

func (a *TestApp) GetConfigFlagSet(fs *flag.FlagSet) {
	log.Logger.Infof("TestApp GetConfigFlagSet")
}

func (a *TestApp) Startup(wg *sync.WaitGroup) {
	log.Logger.Infof("TestApp Startup")
	a.wg = wg
}

func (a *TestApp) Shutdown() {
	log.Logger.Infof("TestApp Shutdown")
}

func (a *TestApp) Check() error {
	log.Logger.Infof("TestApp Check")
	return a.err
}

func TestApplicationRunner_StartStop(t *testing.T) {
	testApp, newAppErr := NewTestApp()
	assert.Nil(t, newAppErr)
	appRunner, err := apprun.NewApplicationRunner(testApp)
	assert.Nil(t, err)

	go func() {
		<-time.After(200 * time.Millisecond)
		// Sent TERM signal
		must.Must(syscall.Kill(syscall.Getpid(), syscall.SIGTERM))
	}()

	// Start the app runner in blocking mode, that will be killed after 200 msec
	appRunner.Run()
}
