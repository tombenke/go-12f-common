package worker

import (
	"context"
	"time"

	"github.com/tombenke/go-12f-common/v2/buildinfo"
	"github.com/tombenke/go-12f-common/v2/examples/scheduler/model"
	"github.com/tombenke/go-12f-common/v2/oti"
	"go.opentelemetry.io/otel/attribute"
)

type ProcessTimerRequestFunc func(ctx context.Context, timerRequest model.TimerRequest) (time.Duration, error)

type ComponentNamer interface {
	ComponentName() string
}

func obsProcessTimerRequest(namer ComponentNamer, observedFn ProcessTimerRequestFunc) ProcessTimerRequestFunc {
	return func(ctx context.Context, timerRequest model.TimerRequest) (time.Duration, error) {
		meter := oti.GetMeter(ctx)
		return oti.Metrics[time.Duration](ctx, meter, "process_request", "Process Request", []attribute.KeyValue{
			oti.FieldApp.String(buildinfo.AppName()),
			oti.FieldService.String(namer.ComponentName()),
		}, oti.FirstErr)(func(ctx context.Context) (time.Duration, error) {
			return observedFn(ctx, timerRequest)
		})(ctx)
	}
}
