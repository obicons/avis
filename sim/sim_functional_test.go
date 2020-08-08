package sim

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/obicons/rmck/util"
)

func TestGazeboFunctional(t *testing.T) {
	gazebo, err := NewGazeboFromEnv()
	if err != nil {
		t.Fatalf("could not get a gazebo from environment: %s", err)
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

	ctx, cc = context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	gazebo.Step(ctx)

	fmt.Println("sleep 30s")
	time.Sleep(30 * time.Second)
}
