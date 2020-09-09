package platforms

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/creack/pty"
	"github.com/obicons/rmck/sim"
	"github.com/obicons/rmck/util"
)

type PX4 struct {
	srcPath string
	cmd     *exec.Cmd
	pty     *os.File
}

func NewPX4FromEnv() (System, error) {
	px4Path := os.Getenv("PX4_PATH")
	if px4Path == "" {
		return nil, fmt.Errorf("error: NewPX4FromEnv(): set PX4_PATH")
	} else if _, err := os.Stat(px4Path); err != nil {
		return nil, fmt.Errorf("error: NewPX4FromEnv(): %s", err)
	}

	px4 := PX4{
		srcPath: px4Path,
		cmd:     nil,
	}

	return &px4, nil
}

// implements System
func (px4 *PX4) Start() error {
	binaryPath := path.Join(px4.srcPath, "bin/px4")
	romfsPath := path.Join(px4.srcPath, "ROMFS/px4fmu_common")
	rcPath := path.Join(px4.srcPath, "etc/init.d-posix/rcS")
	testDataPath := path.Join(px4.srcPath, "test_data")
	rootFs := path.Join(px4.srcPath, "tmp/rootfs")
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("error: Start(): build px4")
	}

	cmd := exec.Command(
		binaryPath,
		"-d", // disable user input
		romfsPath,
		"-s", // set startup path
		rcPath,
		"-t", // set test data
		testDataPath,
	)
	cmd.Dir = rootFs
	cmd.Env = px4Environ()
	// cmd.Stdin = os.Stdin

	logging, err := util.GetLogger("px4")
	if err != nil {
		return err
	}

	// err = util.LogProcess(cmd, logging)
	// if err != nil {
	// 	return err
	// }

	// err = cmd.Start()
	px4.pty, err = pty.Start(cmd)
	if err == nil {
		px4.cmd = cmd
	}

	util.LogReader(px4.pty, logging)

	return err
}

// implements System
func (px4 *PX4) Stop(ctx context.Context) error {
	return util.GracefulStop(px4.cmd, ctx)
}

// implements System
func (px4 *PX4) GetGazeboConfig() (*sim.GazeboConfig, error) {
	worldfilePath := path.Join(px4.srcPath, "Tools/sitl_gazebo/worlds/iris.world")
	if _, err := os.Stat(worldfilePath); err != nil {
		return nil, fmt.Errorf("GetGazeboConfig(): %s", err)
	}

	pluginPath := path.Join(px4.srcPath, "build_gazebo")
	modelPath := path.Join(px4.srcPath, "Tools/sitl_gazebo/models")
	ldLibraryPath := path.Join(px4.srcPath, "build_gazebo")

	conf := sim.GazeboConfig{
		WorldPath: worldfilePath,
		Env: []string{
			fmt.Sprintf("GAZEBO_PLUGIN_PATH=%s", pluginPath),
			fmt.Sprintf("GAZEBO_MODEL_PATH=%s", modelPath),
			fmt.Sprintf("LD_LIBRARY_PATH=%s", ldLibraryPath),
		},
		WorkDir:  px4.srcPath,
		StepSize: 4000000,
	}
	return &conf, nil
}

// returns environment variables needed by PX4
func px4Environ() []string {
	env := os.Environ()
	env = append(
		env,
		"HEADLESS=1",
		"PX4_HOME_LAT=-35.363261",
		"PX4_HOME_LON=149.165230",
		"PX4_HOME_ALT=584",
		"DISPLAY=:0",
		"PX4_SIM_MODEL=iris",
	)
	return env
}
