package apprun

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/tombenke/go-12f-common/v2/config"
	"github.com/tombenke/go-12f-common/v2/gsd"
	"github.com/tombenke/go-12f-common/v2/healthcheck"
	"github.com/tombenke/go-12f-common/v2/log"
	"github.com/tombenke/go-12f-common/v2/oti"
	"go.uber.org/multierr"
)

type HealthCheckHook interface {
	// Called by the healthcheck API. If this function returns with nil that means the component is healthy.
	// If it returns with any error, that means the component is either sick or yet not ready for working.
	Check(ctx context.Context) error
}

// Interface that defines a component's life-cycle management functions.
type ComponentLifecycleManager interface {
	HealthCheckHook

	// Components should initializes itself in this method
	Startup(ctx context.Context, wg *sync.WaitGroup) error

	// Should shut down the component and free all it's resources
	Shutdown(ctx context.Context) error
}

// Interface that defines an application
type Application interface {
	// Should return all component
	Components(ctx context.Context) []ComponentLifecycleManager
}

type AfterStartupHook interface {
	// Hook that runs after the components have been started
	AfterStartup(ctx context.Context, wg *sync.WaitGroup) error
}

type BeforeShutdownHook interface {
	// Hook that runs before components are going to be shut down
	BeforeShutdown(ctx context.Context) error
}

// MakeAndRun() is a wrapper function to make and run an application via ApplicationRunner
func MakeAndRun[T config.Configurer](appConfig T, appFactory func(T) (Application, error)) error {
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
		log.SetupDefault(config.LogLevel, config.LogFormat)

		app, err := appFactory(appConfig)
		if err != nil {
			return fmt.Errorf("failed to create application. %w", err)
		}
		appRunner := NewApplicationRunner(config, app)
		return appRunner.Run()
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
	app    Application
	wg     *sync.WaitGroup
}

// NewApplicationRunner creates a new ApplicationRunner instance
func NewApplicationRunner(config *Config, app Application) *ApplicationRunner {
	return &ApplicationRunner{
		config: config,
		app:    app,
		wg:     &sync.WaitGroup{},
	}
}

// Run() runs the application, that means it calls the Startup() method of the application instance,
// and steps into the execution loop, that runs until the application receives signal to shut it down.
func (ar *ApplicationRunner) Run() error {
	// Initialize the config structures of the runner and the application using default values, envirnonment variables and CLI arguments
	ctx, logger := log.With(context.Background(), "appId", uuid.NewString())

	if logger.Enabled(ctx, slog.LevelDebug) {
		logger.Debug("Starting 12f application", "config", ar.config)
	} else {
		logger.Info("Starting 12f application")
	}
	ar.wg.Add(1)

	// Start the liveness and readiness check
	hc := healthcheck.NewHealthCheck(
		ar.wg,
		healthcheck.Config{
			Port: uint(ar.config.HealthCheckPort),
			Checks: map[string]healthcheck.Check{
				ar.config.LivenessCheckPath:  ar.livenessCheck,
				ar.config.ReadinessCheckPath: ar.readinessCheck,
			},
		},
	)

	// Setup the OTEL instrumentation
	oti := oti.NewOtel(ar.wg, ar.config.OtelConfig)
	ctx = oti.Startup(ctx)

	// Start the startup process of the application to run
	hc.Startup(ctx)

	// Startup every component
	if err := ar.startupComponents(ctx); err != nil {
		return fmt.Errorf("failed to start application components: %w", err)
	}

	if err := ar.waitUntilComponentsAreHealthy(ctx); err != nil {
		return err
	}

	if afterStartupHook, ok := ar.app.(AfterStartupHook); ok {
		if err := afterStartupHook.AfterStartup(ctx, ar.wg); err != nil {
			return fmt.Errorf("after startup hook returned error. %w", err)
		}
	}

	// Setup graceful shutdown
	gsd.RegisterGsdCallback(ctx, ar.wg, func(s os.Signal) {
		defer ar.wg.Done()

		// Shuts down the application
		logger.Info("GsdCallback called")

		// Executes the BeforeShutdown hook if provided
		if beforeShutdownHook, ok := ar.app.(BeforeShutdownHook); ok {
			if err := beforeShutdownHook.BeforeShutdown(ctx); err != nil {
				logger.Error("BeforeShutdown hook returned with error", "error", err)
			}
		}
		// Executes the shutdown process of the application
		if err := ar.shutdownComponents(ctx); err != nil {
			logger.Error("Failed to shut down application", "error", err)
		}

		// Shut down the OTEL services
		oti.Shutdown(ctx)

		// Shut down the healthcheck services
		hc.Shutdown(ctx)
	})

	// Wait until the application has shut down
	ar.wg.Wait()
	return nil
}

// Check components health
func (ar *ApplicationRunner) waitUntilComponentsAreHealthy(ctx context.Context) error {
	// TODO: Make this configurable?
	policy := retrypolicy.NewBuilder[any]().
		WithMaxRetries(-1).
		WithBackoff(25*time.Millisecond, 500*time.Millisecond).
		WithMaxDuration(10 * time.Second).
		Build()

	if err := failsafe.With(policy).Run(func() error {
		var errs error
		for _, c := range ar.app.Components(ctx) {
			multierr.AppendInto(&errs, c.Check(ctx))
		}
		return errs
	}); err != nil {
		return fmt.Errorf("one or more components are not healthy. %w", err)
	}
	return nil
}

func (ar *ApplicationRunner) startupComponents(ctx context.Context) error {
	var err error
	for _, c := range ar.app.Components(ctx) {
		multierr.AppendInto(&err, c.Startup(ctx, ar.wg))
	}
	return err
}

func (ar *ApplicationRunner) shutdownComponents(ctx context.Context) error {
	var err error
	for _, c := range slices.Backward(ar.app.Components(ctx)) {
		multierr.AppendInto(&err, c.Shutdown(ctx))
	}
	return err
}

// livenessCheck() is the built-in livenessCheck callback function for the HealthCheck service
func (ar *ApplicationRunner) livenessCheck(ctx context.Context) error {
	// TODO: May add checks for heap-size, go routine num limit, etc.
	log.DebugContext(ctx, "Liveness check")
	return nil
}

// readinessCheck() is the built-in readinessCheck callback function for the HealthCheck service
func (ar *ApplicationRunner) readinessCheck(ctx context.Context) error {
	var err error
	for _, c := range ar.app.Components(ctx) {
		multierr.AppendInto(&err, c.Check(ctx))
	}
	if healthCheckHook, ok := ar.app.(HealthCheckHook); ok {
		multierr.AppendInto(&err, healthCheckHook.Check(ctx))
	}
	log.DebugContext(ctx, "Readiness check", "error", err)
	return err
}
