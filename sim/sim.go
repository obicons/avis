package sim

import (
	"context"
	"time"

	"github.com/obicons/rmck/entities"
)

type StepActions func()

type Sim interface {
	Start() error
	Shutdown(ctx context.Context) error
	Step(ctx context.Context) error
	SimTime(ctx context.Context) (time.Time, error)
	Position(ctx context.Context) (entities.Position, error)
	AddPostStepAction(action StepActions)
	Iterations() uint64
}
