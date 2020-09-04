package platforms

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/obicons/rmck/sim"
	"github.com/obicons/rmck/util"
)

type ArduPilot struct {
	srcPath         string
	gazeboSrcPath   string
	droneSignalPath string
	cmd             *exec.Cmd
	logger          *log.Logger
	lastMsgTime     time.Time
}

const droneSignalTimeout = time.Millisecond * 250

func NewArduPilotFromEnv() (System, error) {
	// get the environment variable
	srcPath := os.Getenv("ARDUPILOT_SRC_PATH")
	if srcPath == "" {
		return nil, fmt.Errorf("error: NewArduPilotFromEnv(): ARDUPILOT_SRC_PATH not set")
	}
	gzPath := os.Getenv("ARDUPILOT_GZ_PATH")
	if gzPath == "" {
		return nil, fmt.Errorf("error: NewArduPilotFromEnv(): ARDUPILOT_GZ_PATH not set")
	}

	// make sure the path exists
	stat, err := os.Stat(srcPath)
	if err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("error: ARDUPILOT_SRC_PATH (%s) must be a dir", srcPath)
	}

	stat, err = os.Stat(gzPath)
	if err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("error: ARDUPILOT_GZ_PATH (%s) must be a dir", gzPath)
	}

	logger, err := util.GetLogger("ArduPilot Controller")
	if err != nil {
		return nil, fmt.Errorf("error: NewArduPilotFromEnv(): %s", err)
	}

	homedir, _ := os.UserHomeDir()
	droneSignalPath := path.Join(homedir, ".drone_signal")

	ardupilot := ArduPilot{
		srcPath:         srcPath,
		gazeboSrcPath:   gzPath,
		droneSignalPath: droneSignalPath,
		logger:          logger,
	}
	return &ardupilot, nil
}

// implements System
func (a *ArduPilot) Start() error {
	workDir := path.Join(a.srcPath, "Tools/autotest/")

	cmd := exec.Command(
		"./sim_vehicle.py", "-f", "gazebo-iris",
		"--vehicle", "ArduCopter", "--console", "--no-rebuild",
	)
	cmd.Dir = workDir
	cmd.Env = os.Environ()

	// why does this keep us from crashing?
	cmd.Stdin = os.Stdin

	logging, err := util.GetLogger("ardupilot")
	if err != nil {
		return err
	}

	err = util.LogProcess(cmd, logging)
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err == nil {
		a.cmd = cmd
	}

	return err
}

// implements System
func (a *ArduPilot) Stop(ctx context.Context) error {
	return util.GracefulStop(a.cmd, ctx)
}

// implements System
func (a *ArduPilot) GetGazeboConfig() (*sim.GazeboConfig, error) {
	config := sim.GazeboConfig{
		WorkDir:         a.gazeboSrcPath,
		WorldPath:       path.Join(a.gazeboSrcPath, "worlds/iris_arducopter_runway.world"),
		PreStepActions:  []sim.StepActions{func() { a.checkDroneSignal(false) }},
		PostStepActions: []sim.StepActions{func() { a.checkDroneSignal(true) }},
	}
	return &config, nil
}

// connects to the signal the drone is broadcasting
func (a *ArduPilot) checkDroneSignal(isPostStep bool) {
	socket, err := net.Dial("unix", a.droneSignalPath)
	if err != nil {
		if time.Now().Sub(a.lastMsgTime) > time.Second {
			a.logger.Printf("checkDroneSignal: %s\n", err)
			a.lastMsgTime = time.Now()
		}
		return
	}
	defer socket.Close()

	if isPostStep {
		resp := make([]byte, 8)
		socket.Read(resp)
	}
}
