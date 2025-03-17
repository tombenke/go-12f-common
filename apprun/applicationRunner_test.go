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
	"github.com/stretchr/testify/suite"
	"github.com/tombenke/go-12f-common/v2/apprun"
	"github.com/tombenke/go-12f-common/v2/log"
	"github.com/tombenke/go-12f-common/v2/must"
	"github.com/tombenke/go-12f-common/v2/oti"
)

type Config struct{}

type AppRunnerSuite struct {
	suite.Suite
	arWG  *sync.WaitGroup
	arCtx context.Context
	oti   oti.Otel
}

func (s *AppRunnerSuite) SetupSuite() {
	s.arWG = &sync.WaitGroup{}
	s.arCtx, _ = log.WithLogger(context.Background(), slog.Default().With(oti.KeyTestSuite, "AppTestSuite"))

	s.oti = oti.NewOtel(s.arWG, oti.Config{
		OtelTracesExporter:  string(oti.TraceExporterTypeConsole),
		OtelMetricsExporter: string(oti.MetricExporterTypeConsole),
	})
	s.oti.Startup(s.arCtx)
}

func (s *AppRunnerSuite) TearDownSuite() {
	s.oti.Shutdown(s.arCtx)
}

func TestAppRunner(t *testing.T) {
	suite.Run(t, new(AppRunnerSuite))
}

type TestApp struct {
}

func NewTestApp() apprun.Application {
	return &TestApp{}
}

func (a *TestApp) Components(ctx context.Context) []apprun.ComponentLifecycleManager {
	return nil
}

func (s *AppRunnerSuite) TestStartStop() {
	t := s.T()
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
