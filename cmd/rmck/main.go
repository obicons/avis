package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/obicons/rmck/detector"
	"github.com/obicons/rmck/entities"
	"github.com/obicons/rmck/executor"
	"github.com/obicons/rmck/hinj"
	"github.com/obicons/rmck/platforms"
	"github.com/obicons/rmck/sim"
)

type workloadInfo struct {
	AutopilotName string
}

var (
	rpcAddr                = flag.String("rpc.addr", getRPCAddr(), "URL of RPC server")
	autopilot              = flag.String("autopilot", "", "Autopilot to test (ardupilot or px4)")
	workloadCmd            = flag.String("workload.cmd", "", "Command of workload (accepts a Go template)")
	workloadTimeoutSeconds = flag.Uint("workload.timeout", 300, "Timeout of workload (seconds)")
)

func main() {
	flag.Parse()

	system := getAutoPilot(*autopilot)

	hinj, err := hinj.NewHINJServer(getHINJAddr())
	if err != nil {
		log.Fatalf("Could not create a new HINJ server: %s\n", err)
	}

	config, _ := system.GetGazeboConfig()

	gazebo, err := sim.NewGazeboFromEnv(config)
	if err != nil {
		log.Fatalf("Could not get a gazebo instance: %s\n", err)
	}

	workloadCmd, err := parseWorkloadTemplate(*autopilot, *workloadCmd)
	if err != nil {
		log.Fatalf("Could not parse workload command: %s\n", err)
	}

	// this is a profiling run
	positionRecorder := detector.NewPositionRecorder()
	var modeChangeTimes []uint64
	recordModeChanges := func(iterations uint64, mode int) {
		modeChangeTimes = append(modeChangeTimes, iterations)
	}

	ex := executor.Executor{
		HINJServer:  hinj,
		Simulator:   gazebo,
		Autopilot:   system,
		WorkloadCmd: workloadCmd,
		Timeout:     time.Duration(*workloadTimeoutSeconds) * time.Second,
		RPCAddr:     *rpcAddr,
		Detectors: []detector.Detector{
			detector.NewTimeoutDetector(time.Duration(*workloadTimeoutSeconds) * time.Second),
			positionRecorder,
			detector.NewFreeFallDetector(),
		},
		ModeChangeHandler: recordModeChanges,
	}
	if err = ex.Execute(); err != nil {
		log.Fatalf("Error executing: %s\n", err)
	}

	doModelChecking(
		hinj,
		gazebo,
		system,
		workloadCmd,
		positionRecorder.(*detector.PositionRecorder).GetPositions(),
		modeChangeTimes,
	)
}

// does the actual checking
func doModelChecking(hinjServer *hinj.HINJServer,
	sim sim.Sim,
	autopilot platforms.System,
	workloadCmd string,
	positions []entities.Position,
	modeChangeTimes []uint64) {

	// The first item of failurePlans is the next scenario we consider and so on.
	// FIFO order.
	var failurePlans [][]executor.FailurePlan

	// Tracks the failure scenarios that we have considered
	consideredScenarios := make(map[uint64]bool)

	enqueueScenarios(modeChangeTimes, &failurePlans, consideredScenarios)
	for len(failurePlans) > 0 {
		// dequeue the next failure plan
		nextFailurePlan := failurePlans[0]
		failurePlans = failurePlans[1:]

		fmt.Println(nextFailurePlan)

		// we will use this information to create new failure plans
		var modeChangeTimes []uint64
		recordModeChanges := func(iterations uint64, mode int) {
			modeChangeTimes = append(modeChangeTimes, iterations)
		}

		ex := executor.Executor{
			HINJServer:  hinjServer,
			Simulator:   sim,
			Autopilot:   autopilot,
			WorkloadCmd: workloadCmd,
			Timeout:     time.Duration(*workloadTimeoutSeconds) * time.Second,
			RPCAddr:     *rpcAddr,
			Detectors: []detector.Detector{
				detector.NewTimeoutDetector(time.Duration(*workloadTimeoutSeconds) * time.Second),
				detector.NewFreeFallDetector(),
			},
			ModeChangeHandler:  recordModeChanges,
			MissionFailurePlan: nextFailurePlan,
		}
		if err := ex.Execute(); err != nil {
			log.Fatalf("Error executing: %s\n", err)
		}

		// enqeueue the same failures of this run, but with the failure time shifted
		var shiftedFailures []executor.FailurePlan
		for _, failure := range nextFailurePlan {
			shiftedFailure := failure
			shiftedFailure.FailureTime += 1
			shiftedFailures = append(shiftedFailures, shiftedFailure)
		}

		// enqueues only if we haven't considered the shifted failure
		if hash, err := hashstructure.Hash(shiftedFailures, nil); err != nil {
			// this should never occur
			panic(err)
		} else if !consideredScenarios[hash] {
			consideredScenarios[hash] = true
		} // otherwise, we don't need to consider the shifted scenario

		enqueueScenarios(modeChangeTimes, &failurePlans, consideredScenarios)
	}

}

// enqueue the new mode changes from this run.
// at each mode transition, we can inject a subset of our failure powerset.
func enqueueScenarios(modeChangeTimes []uint64, plans *[][]executor.FailurePlan, consideredScenarios map[uint64]bool) {
	for _, modeTimestamp := range modeChangeTimes {
		failures := allFailures(modeTimestamp)
		candidates := failurePowerset(failures)
		// remove scenarios that we have:
		//   i. already considered (hash the scenario and compare)
		//  ii. are not feasible (e.g. redundant failures)
		for _, c := range candidates {
			hash, err := hashstructure.Hash(c, nil)
			if err != nil {
				// should never happen
				panic(err)
			} else if len(c) == 0 {
				continue
			} else if consideredScenarios[hash] {
				continue
			} else if !scenarioFeasible(c) {
				continue
			}
			// the scenario is new and feasible, so enqueue
			consideredScenarios[hash] = true
			*plans = append(*plans, c)
		}
	}
}

// returns if the failure scenario itself is not feasible
func scenarioFeasible(scenario []executor.FailurePlan) bool {
	for i, failure := range scenario {
		count := 0
		for j, other := range scenario {
			if failure.SensorFailure.SensorType == other.SensorFailure.SensorType &&
				failure.SensorFailure.Instance == other.SensorFailure.Instance &&
				i != j {
				return false
			} else if failure.SensorFailure.SensorType == other.SensorFailure.SensorType {
				count++
			}
		}
		if count != 3 {
			return false
		}
	}
	return true
}

// returns all failures at iteration
func allFailures(iteration uint64) []executor.FailurePlan {
	var failures []executor.FailurePlan
	sensorTypes := []hinj.Sensor{hinj.GPS, hinj.Accelerometer}
	for _, sensorType := range sensorTypes {
		// TODO - probably use a map to store # of instances
		for instance := uint8(0); instance < uint8(3); instance++ {
			failures = append(
				failures,
				executor.FailurePlan{
					SensorFailure: hinj.SensorFailure{
						SensorType: sensorType,
						Instance:   instance,
					},
					FailureTime: iteration,
				},
			)
		}
	}
	return failures
}

// returns the powerset of the given failure plan (e.g. all possible failures)
func failurePowerset(failures []executor.FailurePlan) [][]executor.FailurePlan {
	if len(failures) == 0 {
		return [][]executor.FailurePlan{{}}
	}

	var results [][]executor.FailurePlan
	thePowerSet := failurePowerset(failures[1:])
	for _, set := range thePowerSet {
		results = append(results, set)
		results = append(results, append(set, failures[0]))
	}

	return results
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

func getAutoPilot(autopilotName string) platforms.System {
	adjustedName := strings.ToLower(autopilotName)
	var sys platforms.System
	var err error
	switch adjustedName {
	case "ardupilot":
		sys, err = platforms.NewArduPilotFromEnv()
	case "px4":
		sys, err = platforms.NewPX4FromEnv()
	case "":
		err = fmt.Errorf("autopilot name not supplied via -autopilot")
	default:
		err = fmt.Errorf("unknown autopilot name: %s", autopilotName)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot create new autopilot: %s\n", err)
		os.Exit(1)
	}

	return sys
}

func parseWorkloadTemplate(autopilotName, workloadCmd string) (string, error) {
	info := workloadInfo{AutopilotName: autopilotName}
	template := template.New("Workload")
	template, err := template.Parse(workloadCmd)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	if err = template.Execute(&builder, &info); err != nil {
		return "", err
	}

	return builder.String(), nil
}
