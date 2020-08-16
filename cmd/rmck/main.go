package main

import (
	"context"
	"log"
	"time"

	"github.com/obicons/rmck/platforms"
	"github.com/obicons/rmck/sim"
)

func main() {
	gazebo, err := sim.NewGazeboFromEnv()
	if err != nil {
		log.Fatalf("Could not get a Gazebo instance: %s", err)
	}

	ardupilot, err := platforms.NewArduPilotFromEnv()
	if err != nil {
		log.Fatalf("Could not get an ArduPilot instance: %s", err)
	}

	err = gazebo.Start()
	if err != nil {
		log.Fatalf("Could not start Gazebo: %s", err)
	}

	time.Sleep(time.Second * 5)

	err = ardupilot.Start()
	if err != nil {
		log.Fatalf("Could not start ArduPilot: %s", err)
	}

	time.Sleep(time.Second * 120)

	ctx, cc := context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	ardupilot.Stop(ctx)

	ctx, cc = context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	gazebo.Stop(ctx)
}
