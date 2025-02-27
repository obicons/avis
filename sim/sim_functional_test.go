package sim

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/obicons/avis/util"
)

func TestFunctionalGazebo(t *testing.T) {
	sim, err := NewGazeboFromEnv(testingConfig(t))
	if err != nil {
		t.Fatalf("could not get a gazebo from environment: %s", err)
	}

	gazebo, ok := sim.(*Gazebo)
	if !ok {
		t.Fatalf("NewGazeboFromEnv() did not return a *Gazebo")
	}

	gazebo.Start()
	if running, err := util.IsRunning("gzserver"); err != nil || !running {
		t.Fatalf("gazebo does not appear to be running: %s", err)
	}

	defer func() {
		ctx, cc := context.WithTimeout(context.Background(), time.Second*5)
		gazebo.Stop(ctx)
		if gazebo.Cmd.ProcessState == nil {
			t.Fatalf("gazebo appears to still be running")
		}
		cc()
	}()

	ctx, cc := context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	readTime, err := gazebo.SimTime(ctx)
	if err != nil {
		t.Fatalf("could not get time: %s", err)
	}

	if readTime.Second() != 0 || readTime.Nanosecond() != 0 {
		t.Fatalf("expected zero time, found: %s", readTime)
	}

	ctx, cc = context.WithTimeout(context.Background(), time.Second)
	defer cc()
	_, err = gazebo.Position(ctx)
	if err != nil {
		t.Fatalf("could not get position: %s", err)
	}

	ctx, cc = context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	gazebo.Step(ctx)

	ctx, cc = context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	err = gazebo.Stop(ctx)
	if err != nil {
		t.Fatalf("gazebo could not stop: %s", err)
	}
}

func testingConfig(t *testing.T) *GazeboConfig {
	t.Helper()
	gazeboPath := os.Getenv("ARDUPILOT_GZ_PATH")
	if gazeboPath == "" {
		t.Fatal("Error: ARDUPILOT_GZ_PATH not set")
	}
	config := GazeboConfig{
		WorldPath: path.Join(gazeboPath, "worlds/iris_arducopter_runway.world"),
		WorkDir:   gazeboPath,
	}
	return &config
}
