package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/obicons/rmck/hinj"
	"github.com/obicons/rmck/platforms"
	"github.com/obicons/rmck/sim"
)

func main() {
	ardupilot, err := platforms.NewArduPilotFromEnv()
	if err != nil {
		log.Fatalf("Could not get an ArduPilot instance: %s\n", err)
	}

	hinj, err := hinj.NewHINJServer(getHINJAddr())
	if err != nil {
		log.Fatalf("Could not create a new HINJ server: %s\n", err)
	}
	if err = hinj.Start(); err != nil {
		log.Fatalf("Error starting HINJ server: %s\n", err)
	}

	config, _ := ardupilot.GetGazeboConfig()

	gazebo, err := sim.NewGazeboFromEnv(config)
	if err != nil {
		log.Fatalf("Could not get a Gazebo instance: %s\n", err)
	}

	err = gazebo.Start()
	if err != nil {
		log.Fatalf("Could not start Gazebo: %s\n", err)
	}

	time.Sleep(time.Second * 5)

	err = ardupilot.Start()
	if err != nil {
		log.Fatalf("Could not start ArduPilot: %s\n", err)
	}

	time.Sleep(time.Second * 45)
	fmt.Println("done sleeping!")

	startTime := time.Now()
	for i := 0; time.Now().Sub(startTime) < time.Second*60; i++ {
		gazebo.Step(context.Background())
		time.Sleep(time.Millisecond * 10)
	}

	ctx, cc := context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	ardupilot.Stop(ctx)

	ctx, cc = context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	gazebo.Stop(ctx)

	hinj.Shutdown()
}

func getHINJAddr() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	path := path.Join(home, ".hardware_controller")
	if _, err := os.Stat(path); err == nil {
		if err = os.Remove(path); err != nil {
			panic(err)
		}
	}

	addr := fmt.Sprintf("unix://%s", path)
	return addr
}
