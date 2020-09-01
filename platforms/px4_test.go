package platforms

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/obicons/rmck/util"
)

func TestUnitNewPX4FromEnvNoPX4Path(t *testing.T) {
	oldPX4Path := os.Getenv("PX4_PATH")
	os.Setenv("PX4_PATH", "")
	if _, err := NewPX4FromEnv(); err == nil {
		t.Fatalf("Expected NewPX4FromEnv() to produce an error")
	}
	os.Setenv("PX4_PATH", oldPX4Path)
}

func TestUnitNewPX4FromEnvInvalidPath(t *testing.T) {
	oldPX4Path := os.Getenv("PX4_PATH")
	os.Setenv("PX4_PATH", "/no/way/this/is/a/path")
	if _, err := NewPX4FromEnv(); err == nil {
		t.Fatalf("Expected NewPX4FromEnv() to produce an error")
	}
	os.Setenv("PX4_PATH", oldPX4Path)
}

func TestFunctionalPX4(t *testing.T) {
	system, err := NewPX4FromEnv()
	if err != nil {
		t.Fatalf("NewPX4FromEnv() returned an unexpected error: %s", err)
	}

	px4, ok := system.(*PX4)
	if !ok {
		t.Fatal("NewPX4FromEnv() expected to produce a PX4")
	}

	if err = px4.Start(); err != nil {
		t.Fatalf("px4.Start() returned an unexpected error: %s", err)
	}

	startTime, isRunning := time.Now(), false
	for !isRunning && time.Now().Sub(startTime) < time.Second*5 {
		isRunning, _ = util.IsRunning("px4")
		time.Sleep(time.Millisecond * 250)
	}
	if !isRunning {
		t.Fatal("PX4 does not appear to have started")
	}

	time.Sleep(time.Second * 10)

	if err = px4.Stop(context.Background()); err != nil {
		t.Fatalf("px4.Stop() returned an unexpected error: %s", err)
	}
}
