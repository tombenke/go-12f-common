package model

import (
	"context"
	"time"
)

// TimerRequest avoids dependency between timer and worker packages
type TimerRequest struct {
	Ctx         context.Context
	CurrentTime time.Time
}
