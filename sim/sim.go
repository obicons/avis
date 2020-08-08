package sim

import (
	"context"
	"time"
)

type Sim interface {
	Start() error
	Stop(ctx context.Context) error
	Step() error
	SimTime(ctx context.Context) (time.Time, error)
}
