package sim

import (
	"context"
	"time"
)

type Sim interface {
	Start() error
	Stop(ctx context.Context) error
	Step(ctx context.Context) error
	SimTime(ctx context.Context) (time.Time, error)
	Position(ctx context.Context) (Position, error)
}

type Position struct {
	X float64
	Y float64
	Z float64
}
