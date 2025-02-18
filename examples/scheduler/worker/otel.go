package worker

import (
	"context"
	"time"

	"github.com/tombenke/go-12f-common/v2/buildinfo"
	"github.com/tombenke/go-12f-common/v2/examples/scheduler/model"
	"github.com/tombenke/go-12f-common/v2/oti"
)

type ProcessTimerRequestFunc func(ctx context.Context, timerRequest model.TimerRequest) (time.Duration, error)

func obsProcessTimerRequest(namer oti.ComponentNamer, observedFn ProcessTimerRequestFunc) ProcessTimerRequestFunc {
	return func(ctx context.Context, timerRequest model.TimerRequest) (time.Duration, error) {
		return oti.Metrics[time.Duration](ctx, "process_request", "Process Request", map[string]string{
			oti.KeyApp:     buildinfo.AppName(),
			oti.KeyService: namer.ComponentName(),
		}, oti.FirstErr)(func(ctx context.Context) (time.Duration, error) {
			return observedFn(ctx, timerRequest)
		})(ctx)
	}
}
