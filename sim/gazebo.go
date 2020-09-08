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
	"time"

	"github.com/creack/pty"
	"github.com/obicons/rmck/util"
)

type StepActions func()

type Gazebo struct {
	ExecutablePath  string
	Config          *GazeboConfig
	Logger          *log.Logger
	Cmd             *exec.Cmd
	TimePath        string
	StepPath        string
	PositionPath    string
	TotalIterations uint64
	PTY             *os.File
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

	logging, err := util.GetLogger("gazebo")
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
	return nil
}

// implements sim.Sim
func (gazebo *Gazebo) Stop(ctx context.Context) error {
	if gazebo.Cmd.ProcessState != nil && gazebo.Cmd.ProcessState.Exited() {
		return fmt.Errorf("Cannot stop gazebo: already exited with status %d", gazebo.Cmd.ProcessState.ExitCode())
	}
	util.GracefulStop(gazebo.Cmd, ctx)
	gazebo.PTY.Close()
	return nil
}

// implements sim.Sim
func (gazebo *Gazebo) SimTime(ctx context.Context) (time.Time, error) {
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
		}
	}
	return gzTime, err
}

/// implements sim.Sim
func (g *Gazebo) Step(ctx context.Context) error {
	var err error
	done := ctx.Done()
	tryToConnect := true
	g.doPreStep()
	for tryToConnect {
		select {
		case <-done:
			err = ctx.Err()
			tryToConnect = false
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

func (g *Gazebo) doPreStep() {
	for _, action := range g.Config.PreStepActions {
		action()
	}
}

func (g *Gazebo) doPostStep() {
	for _, action := range g.Config.PostStepActions {
		action()
	}
}

func NewGazeboFromEnv(config *GazeboConfig) (*Gazebo, error) {
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
	return gazebo, nil
}
