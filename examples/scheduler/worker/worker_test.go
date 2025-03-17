package worker

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/tombenke/go-12f-common/v2/examples/scheduler/model"
	"github.com/tombenke/go-12f-common/v2/log"
	"github.com/tombenke/go-12f-common/v2/oti"
)

type WorkerTestSuite struct {
	suite.Suite
	arWG  *sync.WaitGroup
	arCtx context.Context
	oti   oti.Otel
}

func (s *WorkerTestSuite) SetupSuite() {
	s.arWG = &sync.WaitGroup{}
	s.arCtx, _ = log.WithLogger(context.Background(), slog.Default().With(oti.KeyTestSuite, "WorkerTestSuite"))

	s.oti = oti.NewOtel(s.arWG, oti.Config{
		OtelTracesExporter:  string(oti.TraceExporterTypeConsole),
		OtelMetricsExporter: string(oti.MetricExporterTypeConsole),
	})
	s.oti.Startup(s.arCtx)
}

func (s *WorkerTestSuite) TearDownSuite() {
	s.oti.Shutdown(s.arCtx)
}

func TestWorker(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}

// Create a Worker
func createWorker(t *testing.T) (*Worker, *sync.WaitGroup) {
	worker := NewWorker(
		&Config{},
		make(chan model.TimerRequest),
	)
	fs := &pflag.FlagSet{}
	worker.GetConfigFlagSet(fs)
	wg := &sync.WaitGroup{}
	require.NoError(t, worker.Startup(context.Background(), wg))
	return worker, wg
}

// Create, Start and stop a worker instance
func (s *WorkerTestSuite) TestWorkerStartStop() {
	t := s.T()
	worker, wg := createWorker(t)
	s.NoError(worker.Shutdown(s.arCtx))
	wg.Wait()
}

func (s *WorkerTestSuite) TestWorkerReceive() {
	t := s.T()
	worker, wg := createWorker(t)
	worker.currentTimeCh <- model.TimerRequest{Ctx: s.arCtx, CurrentTime: time.Now()}
	s.NoError(worker.Shutdown(s.arCtx))
	wg.Wait()
}
