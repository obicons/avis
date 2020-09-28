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
	mavproxy        *exec.Cmd
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
	err := a.startArduPilot()
	if err != nil {
		return err
	}

	err = a.startMAVProxy()
	if err != nil {
		a.cmd.Process.Kill()
		a.cmd.Process.Wait()
	}

	return err
}

func (a *ArduPilot) startArduPilot() error {
	workDir := path.Join(a.srcPath)
	defaultsFlag := path.Join(a.srcPath, "Tools/autotest/default_params/copter.parm") +
		"," + path.Join(a.srcPath, "Tools/autotest/default_params/gazebo-iris.parm")

	cmd := exec.Command(
		path.Join(a.srcPath, "build/sitl/bin/arducopter"),
		"-S",
		"-I0",
		"--home",
		"-35.363261,149.165230,584,353",
		"--model",
		"gazebo-iris",
		"--speedup",
		"1",
		"--defaults",
		defaultsFlag,
	)
	cmd.Dir = workDir
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin

	logging, err := util.GetLogger("ardupilot ")
	if err != nil {
		return err
	}

	if err = util.LogProcess(cmd, logging); err != nil {
		return err
	}

	a.cmd = cmd

	return cmd.Start()
}

func (a *ArduPilot) startMAVProxy() error {
	workDir := path.Join(a.srcPath, "Tools/autotest/")

	cmd := exec.Command(
		"./mavproxy.py",
		"--master",
		"tcp:127.0.0.1:5760",
		"--sitl",
		"127.0.0.1:5501",
		"--out",
		"127.0.0.1:14550",
		"--out",
		"127.0.0.1:14551",
		"--console",
	)
	cmd.Dir = workDir
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin

	logging, err := util.GetLogger("mavproxy ")
	if err != nil {
		return err
	}

	if err = util.LogProcess(cmd, logging); err != nil {
		return err
	}

	a.mavproxy = cmd

	return cmd.Start()
}

// implements System
func (a *ArduPilot) Shutdown(ctx context.Context) error {
	firstErr := a.cmd.Process.Kill()
	secondErr := a.cmd.Wait()
	thirdErr := a.mavproxy.Process.Kill()
	fourthErr := a.mavproxy.Wait()

	if firstErr != nil {
		return firstErr
	} else if secondErr != nil {
		return secondErr
	} else if thirdErr != nil {
		return thirdErr
	} else if fourthErr != nil {
		return fourthErr
	}

	return nil
}

// implements System
func (a *ArduPilot) GetGazeboConfig() (*sim.GazeboConfig, error) {
	config := sim.GazeboConfig{
		WorkDir:         a.gazeboSrcPath,
		WorldPath:       path.Join(a.gazeboSrcPath, "worlds/iris_arducopter_runway.world"),
		PreStepActions:  []sim.StepActions{func() { a.checkDroneSignal(false) }},
		PostStepActions: []sim.StepActions{func() { a.checkDroneSignal(true) }},
		StepSize:        1000000,
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
