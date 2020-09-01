package platforms

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/obicons/rmck/util"
)

type ArduPilot struct {
	srcPath string
	cmd     *exec.Cmd
}

func NewArduPilotFromEnv() (*ArduPilot, error) {
	// get the environment variable
	srcPath := os.Getenv("ARDUPILOT_SRC_PATH")
	if srcPath == "" {
		return nil, fmt.Errorf("error: ARDUPILOT_SRC_PATH not set")
	}

	// make sure the path exists
	stat, err := os.Stat(srcPath)
	if err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil,
			fmt.Errorf(
				"error: ARDUPILOT_SRC_PATH (%s) must be a dir",
				srcPath,
			)
	}

	ardupilot := ArduPilot{srcPath: srcPath}
	return &ardupilot, nil
}

// implements System
func (a *ArduPilot) Start() error {
	workDir := path.Join(a.srcPath, "Tools/autotest/")

	// TODO: is --sim-port-in still necessary?
	cmd := exec.Command(
		"./sim_vehicle.py", "-f", "gazebo-iris",
		"--vehicle", "ArduCopter", "--console", "--no-rebuild",
		"--sitl-instance-args", "--sim-port-in", "10000",
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
