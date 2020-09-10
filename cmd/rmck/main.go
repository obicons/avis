package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/obicons/rmck/controller"
	"github.com/obicons/rmck/hinj"
	"github.com/obicons/rmck/platforms"
	"github.com/obicons/rmck/sim"
)

var (
	rpcAddr = flag.String("rpc.addr", getRPCAddr(), "URL of RPC server")
)

func main() {
	flag.Parse()

	ardupilot, err := platforms.NewArduPilotFromEnv()
	if err != nil {
		log.Fatalf("Could not get an ArduPilot instance: %s\n", err)
	}

	// px4, err := platforms.NewPX4FromEnv()
	// if err != nil {
	// 	log.Fatalf("Could not get a PX4 instance: %s\n", err)
	// }

	hinj, err := hinj.NewHINJServer(getHINJAddr())
	if err != nil {
		log.Fatalf("Could not create a new HINJ server: %s\n", err)
	}
	if err = hinj.Start(); err != nil {
		log.Fatalf("Error starting HINJ server: %s\n", err)
	}

	config, _ := ardupilot.GetGazeboConfig()
	// config, _ := px4.GetGazeboConfig()

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
	// err = px4.Start()
	// if err != nil {
	// 	log.Fatalf("Could not start PX4: %s\n", err)
	// }

	fmt.Println("sleeping")
	time.Sleep(time.Second * 35)
	fmt.Println("done sleeping!")

	rpcServer, err := controller.New(*rpcAddr, gazebo)
	if err != nil {
		log.Fatalf("Could not create RPC server: %s\n", err)
	}
	go func() {
		err := rpcServer.Start()
		if err != nil {
			log.Fatalf("Could not start RPC server: %s\n", err)
		}
	}()

	fmt.Println("sleeping for a LONG time!")
	time.Sleep(time.Hour)

	startTime := time.Now()
	for i := 0; time.Now().Sub(startTime) < time.Second*60; i++ {
		err = gazebo.Step(context.Background())
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("done stepping!")

	ctx, cc := context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	ardupilot.Stop(ctx)
	// px4.Stop(ctx)

	ctx, cc = context.WithTimeout(context.Background(), time.Second*5)
	defer cc()
	gazebo.Stop(ctx)

	hinj.Shutdown()
	rpcServer.Stop()
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

func getRPCAddr() string {
	path := os.ExpandEnv("$HOME/.rmck_rpc")
	if _, err := os.Stat(path); err == nil {
		if err = os.Remove(path); err != nil {
			panic(err)
		}
	}
	return "unix://" + path
}
