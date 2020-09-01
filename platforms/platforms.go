package platforms

import (
	"context"

	"github.com/obicons/rmck/sim"
)

type System interface {
	// starts the autopilot
	Start() error

	// stops the autopilot
	Stop(ctx context.Context) error

	// Gets the gazebo configuration.
	// If gazebo is unsupported, return an error.
	GetGazeboConfig() (*sim.GazeboConfig, error)
}
