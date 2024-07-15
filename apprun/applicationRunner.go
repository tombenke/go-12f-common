package apprun

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tombenke/go-12f-common/gsd"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"
)

// LifecycleManager is an interface that defines the application's life-cycle management functions.
// Every application must implement this interface that we want to run via ApplicationRunner
type LifecycleManager interface {
	// Startup() Starts the application and its internal components, by calling their Startup() method.
	Startup(wg *sync.WaitGroup) error

	// Shutdown() Shuts down the internal components, by calling their Shutdown() method, then shuts down the application as well.
	Shutdown() error

	// Check() is called by the healthcheck API. If this function returns with nil that means the appication or component is healthy.
	// If it returns with any error, that means the application or component is either sick, or yet not ready for working.
	Check() error
}

// MakeAndRun() is a wrapper function to make and run an application via ApplicationRunner
func MakeAndRun[T Configurer](appConfig T, appFactory func(T) (LifecycleManager, error)) error {
	rootCmd := &cobra.Command{}
	config := &Config{}
	config.GetConfigFlagSet(rootCmd.Flags())
	appConfig.GetConfigFlagSet(rootCmd.Flags())

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(cmd.Flags()); err != nil {
			return err
		}
		if err := appConfig.LoadConfig(cmd.Flags()); err != nil {
			return err
		}

		app, err := appFactory(appConfig)
		if err != nil {
			return fmt.Errorf("failed to create application. %w", err)
		}
		appRunner := NewApplicationRunner(config, app)
		appRunner.Run()
		return nil
	}
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute command. %w", err)
	}
	return nil
}

// ApplicationRunner is the object, that holds the application,
// and all the supporting components that are needed for a 12-factor application
type ApplicationRunner struct {
	config *Config
	app    LifecycleManager
	wg     *sync.WaitGroup
}

// NewApplicationRunner creates a new ApplicationRunner instance
func NewApplicationRunner(config *Config, app LifecycleManager) *ApplicationRunner {
	return &ApplicationRunner{
		config: config,
		app:    app,
		wg:     &sync.WaitGroup{},
	}
}

// Run() runs the application, that means it calls the Startup() method of the application instance,
// and steps into the execution loop, that runs until the application receives signal to shut it down.
func (ar *ApplicationRunner) Run() {
	// Initialize the config structures of the runner and the application using default values, envirnonment variables and CLI arguments
	log.SetLevelStr(ar.config.LogLevel)
	log.SetFormatterStr(ar.config.LogFormat)
	log.Logger.Debugf("ar.config: %+v", ar.config)
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
	must.Must(ar.app.Startup(ar.wg))

	// Setup graceful shutdown
	gsd.RegisterGsdCallback(ar.wg, func(s os.Signal) {
		defer ar.wg.Done()

		// Shuts down the application
		log.Logger.Infof("ApplicationRunner GsdCallback called")

		// Execute the shutdown process of the application
		if err := ar.app.Shutdown(); err != nil {
			log.Logger.Errorf("Failed to shut down application. %v", err)
		}

		// Shut down the healthcheck services
		hc.Shutdown()
	})

	// Wait until the application has shut down
	ar.wg.Wait()
}

// livenessCheck() is the built-in livenessCheck callback function for the HealthCheck service
func (ar *ApplicationRunner) livenessCheck() error {
	// TODO: May add checks for heap-size, go routine num limit, etc.
	return nil
}

// readinessCheck() is the built-in readinessCheck callback function for the HealthCheck service
func (ar *ApplicationRunner) readinessCheck() error {
	return ar.app.Check()
}
