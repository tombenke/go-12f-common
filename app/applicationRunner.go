package app

import (
	"github.com/tombenke/go-12f-common/cli"
	"github.com/tombenke/go-12f-common/gsd"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"
	"os"
	"sync"
)

// Wrapper function to make and run an application via ApplicationRunner
func MakeAndRun(appWg *sync.WaitGroup, appFactory func(wg *sync.WaitGroup) (LifecycleManager, error)) {
	exampleApp := must.MustVal(appFactory(appWg))
	must.MustVal(NewApplicationRunner(appWg, exampleApp)).Run()
}

// The ApplicationRunner object, that holds the application,
// and all the supporting components that are needed for a 12-factor application
type ApplicationRunner struct {
	config Config
	wg     *sync.WaitGroup
	app    LifecycleManager
}

// Create a new ApplicationRunner instance
func NewApplicationRunner(wg *sync.WaitGroup, app LifecycleManager) (*ApplicationRunner, error) {

	applicationRunner := &ApplicationRunner{wg: wg, app: app, config: Config{}}

	return applicationRunner, nil
}

// Run the application
func (ar *ApplicationRunner) Run() chan os.Signal {

	cli.InitConfigs(os.Args, []cli.FlagSetFunc{ar.config.GetConfigFlagSet, ar.app.GetConfigFlagSet})
	////log.Logger.Infof("ar.config: %v", ar.config)
	log.SetLevelStr(ar.config.LogLevel)
	log.Logger.Infof("ApplicationRunner Run")
	ar.wg.Add(1)

	hc := must.MustVal(healthcheck.NewHealthCheck(ar.wg, healthcheck.Config{Port: uint(ar.config.HealthCheckPort), Checks: map[string]healthcheck.Check{
		ar.config.LivenessCheckPath:  ar.livenessCheck,
		ar.config.ReadinessCheckPath: ar.readinessCheck,
	}}))
	hc.Startup()

	// Start the application
	ar.app.Startup()

	// Setup graceful shutdown
	return gsd.RegisterGsdCallback(ar.wg, func(s os.Signal) {
		// Shuts down the application
		log.Logger.Infof("ApplicationRunner GsdCallback called")
		ar.app.Shutdown()
		hc.Shutdown()
		ar.wg.Done()
	})
}

// The built-in livenessCheck callback function for the HealthCheck service
func (ar *ApplicationRunner) livenessCheck() error {
	// TODO: May add checks for heap-size, go routine num limit, etc.
	return nil
}

// The built-in readinessCheck callback function for the HealthCheck service
func (ar *ApplicationRunner) readinessCheck() error {
	return ar.app.Check()
}
