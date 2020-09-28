package sim

import (
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/obicons/rmck/entities"
	"github.com/obicons/rmck/util"
)

type Gazebo struct {
	sync.Mutex
	ExecutablePath           string
	Config                   *GazeboConfig
	Logger                   *log.Logger
	Cmd                      *exec.Cmd
	TimePath                 string
	StepPath                 string
	PositionPath             string
	TotalIterations          uint64
	PTY                      *os.File
	TemporaryPostStepActions []StepActions
	lastTimeUpdate           int
	lastTime                 time.Time
}

type GazeboConfig struct {
	// contains the world configuration
	WorldPath string

	// where to execute Gazebo
	WorkDir string

	// Any additional environment variables
	Env []string

	// Actions to invoke pre-step
	PreStepActions []StepActions

	// Actions to invoke post-step
	PostStepActions []StepActions

	// Length of each unit of simulation
	StepSize uint64
}

// implements sim.Sim
func (gazebo *Gazebo) Start() error {
	var cmd *exec.Cmd
	cmd = exec.Command(gazebo.ExecutablePath, "--pause", "--verbose", gazebo.Config.WorldPath)
	cmd.Dir = gazebo.Config.WorkDir
	cmd.Env = append(os.Environ(), []string{"DISPLAY=:0", "LC_ALL=C"}...)
	cmd.Env = append(cmd.Env, gazebo.Config.Env...)

	logging, err := util.GetLogger("gazebo ")
	if err != nil {
		return err
	}

	gazebo.Logger = logging

	pty, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	gazebo.PTY = pty
	util.LogReader(pty, logging)
	gazebo.Cmd = cmd
	gazebo.TotalIterations = 0
	return nil
}

// implements sim.Sim
func (gazebo *Gazebo) Shutdown(ctx context.Context) error {
	if gazebo.Cmd.ProcessState != nil && gazebo.Cmd.ProcessState.Exited() {
		return fmt.Errorf("Cannot stop gazebo: already exited with status %d", gazebo.Cmd.ProcessState.ExitCode())
	}
	util.GracefulStop(gazebo.Cmd, ctx)
	gazebo.PTY.Close()
	gazebo.TemporaryPostStepActions = []StepActions{}

	// resets the time cache
	gazebo.lastTimeUpdate = -1

	return nil
}

// implements sim.Sim
// safe to call in multi-threaded environments.
func (gazebo *Gazebo) SimTime(ctx context.Context) (time.Time, error) {
	gazebo.Lock()
	defer gazebo.Unlock()

	// check if we already had this value
	if foundInCache, time := gazebo.checkTimeCache(); foundInCache {
		return time, nil
	}

	done := ctx.Done()
	tryToConnect := true
	coolOffPeriod := 10 * time.Millisecond
	var gzTime time.Time
	var err error
	for tryToConnect {
		select {
		case <-done:
			err = ctx.Err()
			tryToConnect = false
		default:
			addr, err := net.Dial("unix", gazebo.TimePath)
			if err != nil {
				time.Sleep(coolOffPeriod)
				continue
			}
			defer addr.Close()

			// once we have been accepted, we demand fast processesing
			var bytes []byte
			addr.SetDeadline(time.Now().Add(time.Millisecond * 100))
			bytes, err = ioutil.ReadAll(addr)
			if err != nil {
				break
			}

			seconds := int64(binary.LittleEndian.Uint64(bytes[0:8]))
			microseconds := int64(binary.LittleEndian.Uint64(bytes[8:16]))
			gzTime = time.Unix(seconds, microseconds*1000)
			tryToConnect = false

			// update the time cache
			gazebo.lastTime = gzTime
			gazebo.lastTimeUpdate = int(gazebo.TotalIterations)
		}
	}
	return gzTime, err
}

/// implements sim.Sim
func (g *Gazebo) Step(ctx context.Context) error {
	var err error
	// done := ctx.Done()
	tryToConnect := true
	g.doPreStep()
	for tryToConnect {
		select {
		// case <-done:
		// 	err = ctx.Err()
		// 	tryToConnect = false
		default:
			addr, err := net.Dial("unix", g.StepPath)
			if err != nil {
				time.Sleep(time.Millisecond * 10)
				continue
			}

			var bytes []byte = make([]byte, 8)
			g.TotalIterations += 1
			binary.LittleEndian.PutUint64(
				bytes,
				uint64(g.TotalIterations*g.Config.StepSize),
			)
			_, err = addr.Write(bytes)
			addr.Close()
			if err != nil {
				break
			}
			err = nil
			tryToConnect = false
		}
	}
	g.doPostStep()
	return err
}

// implements sim.Sim
func (g *Gazebo) Position(ctx context.Context) (entities.Position, error) {
	position := entities.Position{}
	done := ctx.Done()
	keepTrying := true
	for keepTrying {
		select {
		case <-done:
			keepTrying = false
		default:
			conn, err := net.Dial("unix", g.PositionPath)
			if err != nil {
				time.Sleep(time.Millisecond)
				continue
			}

			var positionBytes [24]byte
			if n, err := conn.Read(positionBytes[:]); err != nil {
				conn.Close()
				time.Sleep(time.Millisecond)
				continue
			} else if n != len(positionBytes) {
				conn.Close()
				time.Sleep(time.Millisecond)
				continue
			}

			// this should never fail (see the test case in sim_test.go)
			conn.Close()
			util.ReadPackedStruct(positionBytes[:], &position)
			keepTrying = false
		}
	}
	return position, ctx.Err()
}

// implements sim.Sim
func (g *Gazebo) AddPostStepAction(action StepActions) {
	g.TemporaryPostStepActions = append(g.TemporaryPostStepActions, action)
}

// implements sim.Sim
// safe to call from a post-step action.
func (g *Gazebo) Iterations() uint64 {
	return g.TotalIterations
}

func (g *Gazebo) checkTimeCache() (bool, time.Time) {
	if g.lastTimeUpdate != -1 && uint64(g.lastTimeUpdate) == g.TotalIterations {
		return true, g.lastTime
	}
	return false, time.Time{}
}

func (g *Gazebo) doPreStep() {
	for _, action := range g.Config.PreStepActions {
		action()
	}
}

func (g *Gazebo) doPostStep() {
	for _, action := range g.Config.PostStepActions {
		action()
	}
	for _, action := range g.TemporaryPostStepActions {
		action()
	}
}

func NewGazeboFromEnv(config *GazeboConfig) (Sim, error) {
	gazeboPath, error := exec.LookPath("gzserver")
	if error != nil {
		return nil, fmt.Errorf("error: gzserver not found on PATH")
	}

	gazebo := new(Gazebo)
	gazebo.ExecutablePath = gazeboPath
	gazebo.Config = config
	gazebo.TimePath = path.Join(os.Getenv("HOME"), ".gazebo_time")
	gazebo.StepPath = path.Join(os.Getenv("HOME"), ".gazebo_world_control")
	gazebo.PositionPath = path.Join(os.Getenv("HOME"), ".gazebo_position")
	gazebo.lastTimeUpdate = -1
	return gazebo, nil
}
