package apprun

import (
	"flag"
	"github.com/tombenke/go-12f-common/cli"
	"github.com/tombenke/go-12f-common/gsd"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"
	"os"
	"sync"
)

// Generic Application life-cycle management functions
// Every application must implement this interface that we want to run via ApplicationRunner
type LifecycleManager interface {
	GetConfigFlagSet(fs *flag.FlagSet)
	Startup(wg *sync.WaitGroup)
	Shutdown()
	Check() error
}

// Wrapper function to make and run an application via ApplicationRunner
func MakeAndRun(appFactory func() (LifecycleManager, error)) {
	exampleApp := must.MustVal(appFactory())
	must.MustVal(NewApplicationRunner(exampleApp)).Run()
}

// The ApplicationRunner object, that holds the application,
// and all the supporting components that are needed for a 12-factor application
type ApplicationRunner struct {
	config Config
	wg     *sync.WaitGroup
	app    LifecycleManager
}

// Create a new ApplicationRunner instance
func NewApplicationRunner(app LifecycleManager) (*ApplicationRunner, error) {

	appWg := &sync.WaitGroup{}
	applicationRunner := &ApplicationRunner{wg: appWg, app: app, config: Config{}}

	return applicationRunner, nil
}

// Run the application
func (ar *ApplicationRunner) Run() {

	// Initialize the config structures of the runner and the application using default values, envirnonment variables and CLI arguments
	cli.InitConfigs(os.Args, []cli.FlagSetFunc{ar.config.GetConfigFlagSet, ar.app.GetConfigFlagSet})
	log.SetLevelStr(ar.config.LogLevel)
	log.SetFormatterStr(ar.config.LogFormat)
	log.Logger.Debugf("ar.config: %v", ar.config)
	log.Logger.Infof("ApplicationRunner Run")
	ar.wg.Add(1)

	// Start the liveness and readiness check
	hc := must.MustVal(healthcheck.NewHealthCheck(ar.wg, healthcheck.Config{Port: uint(ar.config.HealthCheckPort), Checks: map[string]healthcheck.Check{
		ar.config.LivenessCheckPath:  ar.livenessCheck,
		ar.config.ReadinessCheckPath: ar.readinessCheck,
	}}))

	// Start the startup process of the application to run
	hc.Startup()

	// Execute the startup process of the application
	ar.app.Startup(ar.wg)

	// Setup graceful shutdown
	gsd.RegisterGsdCallback(ar.wg, func(s os.Signal) {
		defer ar.wg.Done()

		// Shuts down the application
		log.Logger.Infof("ApplicationRunner GsdCallback called")

		// Execute the shutdown process of the application
		ar.app.Shutdown()

		// Shut down the healthcheck services
		hc.Shutdown()
	})

	// Wait until the application has shut down
	ar.wg.Wait()
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
