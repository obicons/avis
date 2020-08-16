package platforms

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/obicons/rmck/util"
)

func TestNewArduPilotFromEnvNoEnvVar(t *testing.T) {
	prev := os.Getenv("ARDUPILOT_SRC_PATH")
	defer os.Setenv("ARDUPILOT_SRC_PATH", prev)

	// the same as having no value
	err := os.Setenv("ARDUPILOT_SRC_PATH", "")
	if err != nil {
		t.Fatalf("Setenv returned unexpected error: %s", err)
	}

	_, err = NewArduPilotFromEnv()
	if err == nil {
		t.Fatal("NewArduPilotFromEnv did not return an error when it should have")
	}
}

func TestNewArduPilotFromEnvBadPath(t *testing.T) {
	prev := os.Getenv("ARDUPILOT_SRC_PATH")
	defer os.Setenv("ARDUPILOT_SRC_PATH", prev)

	err := os.Setenv("ARDUPILOT_SRC_PATH", "/there/is/no/way/this/path/exists")
	if err != nil {
		t.Fatalf("Setenv returned unexpected error: %s", err)
	}

	_, err = NewArduPilotFromEnv()
	if err == nil {
		t.Fatal("NewArduPilotFromEnv did not return an error when it should have")
	}
}

func TestArduPilotFunctional(t *testing.T) {
	ardupilot, err := NewArduPilotFromEnv()
	if err != nil {
		t.Fatalf("NewArduPilotFromEnv returned an unexpected error: %s", err)
	}

	err = ardupilot.Start()
	if err != nil {
		t.Fatalf("Start returned an unexpected error: %s", err)
	}

	startTime := time.Now()
	isRunning := false
	for !isRunning && time.Now().Sub(startTime) < time.Second*5 {
		isRunning, _ = util.IsRunning("arducopter")
	}
	if !isRunning {
		t.Fatal("ArduCopter does not appear to have started")
	}

	ctx, cc := context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	ardupilot.Stop(ctx)

	if !ardupilot.cmd.ProcessState.Exited() {
		t.Fatal("ArduCopter did not successfully stop")
	}
}
