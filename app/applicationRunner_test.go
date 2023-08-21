package app_test

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/app"
	"github.com/tombenke/go-12f-common/log"
	"sync"
	"syscall"
	"testing"
)

type TestApp struct {
	wg  *sync.WaitGroup
	err error
}

func (a *TestApp) GetConfigFlagSet(fs *flag.FlagSet) {
	log.Logger.Infof("TestApp GetConfigFlagSet")
}

func (a *TestApp) Startup() {
	log.Logger.Infof("TestApp Startup")
}

func (a *TestApp) Shutdown() {
	log.Logger.Infof("TestApp Shutdown")
}

func (a *TestApp) Check() error {
	log.Logger.Infof("TestApp Check")
	return a.err
}

func TestApplicationRunner_StartStop(t *testing.T) {
	appWg := sync.WaitGroup{}
	appRunner, err := app.NewApplicationRunner(&appWg, &TestApp{wg: &appWg})
	assert.Nil(t, err)
	shutdownSigCh := appRunner.Run()

	// Sent TERM signal, then wait for termination
	shutdownSigCh <- syscall.SIGTERM

	// Wait until the application has shut down
	appWg.Wait()

	//TODO: assert.True(t, gsdCbCalled)
}
