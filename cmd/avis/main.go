package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/obicons/avis/detector"
	"github.com/obicons/avis/entities"
	"github.com/obicons/avis/executor"
	"github.com/obicons/avis/hinj"
	"github.com/obicons/avis/platforms"
	"github.com/obicons/avis/sim"
)

type workloadInfo struct {
	AutopilotName string
}

type stats struct {
	totalUnsafe       uint
	unsafeFromGPS     uint
	unsafeFromBaros   uint
	unsafeFromAccel   uint
	unsafeFromCompass uint
	unsafeFromGyro    uint
}

var (
	rpcAddr                       = flag.String("rpc.addr", getRPCAddr(), "URL of RPC server")
	autopilot                     = flag.String("autopilot", "", "Autopilot to test (ardupilot or px4)")
	workloadCmd                   = flag.String("workload.cmd", "", "Command of workload (accepts a Go template)")
	workloadTimeoutSeconds        = flag.Uint("workload.timeout", 300, "Timeout of workload (seconds)")
	inReplay                      = flag.Bool("replay", false, "Perform a replay (requires replay.path to be setup)")
	replayPath                    = flag.String("replay.path", "", "Path to a file containing a trace to replay")
	outputLocation                = flag.String("output", getOutputLocation(), "")
	doSensorTrace                 = flag.Bool("sensor.trace", false, "record the outputs of sensors")
	accelOutputLocation           = flag.String("sensor.accel.output", getSensorOutputLocation("accel"), "")
	gpsOutputLocation             = flag.String("sensor.gps.output", getSensorOutputLocation("gps"), "")
	gyroOutputLocation            = flag.String("sensor.gyro.output", getSensorOutputLocation("gyro"), "")
	compassOutputLocation         = flag.String("sensor.compass.output", getSensorOutputLocation("compass"), "")
	barometerOutputLocation       = flag.String("sensor.barometer.output", getSensorOutputLocation("barometer"), "")
	repl                          = flag.Bool("repl", false, "launch program in REPL mode (does no checking; runs vehicle + hinj)")
	modeOutputDirectory           = flag.String("sensor.mode.output", getSensorOutputLocation("mode"), "")
	signals                       = make(chan os.Signal, 1)
	statistics              stats = stats{}
)

func main() {
	flag.Parse()
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	if *inReplay {
		if *replayPath == "" {
			fmt.Fprintf(os.Stderr, "error: -replay.path must be specified with -replay.\n")
			os.Exit(1)
		} else if info, err := os.Stat(*replayPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		} else if info.IsDir() {
			fmt.Fprintf(os.Stderr, "error: %s is not a file\n", *replayPath)
			os.Exit(1)
		}
		performReplay()
	} else if *repl {
		performREPL()
	} else {
		if _, err := os.Stat(*outputLocation); err != nil {
			os.Mkdir(*outputLocation, 0777)
		}
		performModelChecking()
	}
}

// called to perform a profile run and start the model checking process
func performModelChecking() {
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
		TraceParameters: entities.SensorTraceParameters{
			TraceSensors:         *doSensorTrace,
			AccelTraceOutput:     *accelOutputLocation,
			GPSTraceOutput:       *gpsOutputLocation,
			GyroTraceOutput:      *gyroOutputLocation,
			CompassTraceOutput:   *compassOutputLocation,
			BarometerTraceOutput: *barometerOutputLocation,
		},
	}

	log.Println("Performing a dry run...")
	doneChan := make(chan int)
	go func() {
		if err = ex.Execute(); err != nil {
			log.Fatalf("Error executing: %s\n", err)
		}
		doneChan <- 1
	}()

	select {
	case <-signals:
		log.Println("Received signal, exiting")
		os.Exit(0)
	case <-doneChan:
		// do nothing, this is normal
	}

	if err := saveModes(modeChangeTimes); err != nil {
		log.Printf("error saving mode transitions: %s", err)
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

// performs a replay
func performReplay() {
	var failurePlan []executor.FailurePlan
	file, err := os.Open(*replayPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&failurePlan); err != nil {
		panic(err)
	}

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

	ex := executor.Executor{
		HINJServer:  hinj,
		Simulator:   gazebo,
		Autopilot:   system,
		WorkloadCmd: workloadCmd,
		Timeout:     time.Duration(*workloadTimeoutSeconds) * time.Second,
		RPCAddr:     *rpcAddr,
		Detectors: []detector.Detector{
			detector.NewTimeoutDetector(time.Duration(*workloadTimeoutSeconds) * time.Second),
			detector.NewFreeFallDetector(),
		},
		ModeChangeHandler:  func(totalIterations uint64, modeNumber int) {},
		MissionFailurePlan: failurePlan,
	}
	if err = ex.Execute(); err != nil {
		panic(err)
	}

}

// launches a REPL
func performREPL() {
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

	ex := executor.Executor{
		HINJServer:        hinj,
		Simulator:         gazebo,
		Autopilot:         system,
		WorkloadCmd:       workloadCmd,
		Timeout:           time.Duration(*workloadTimeoutSeconds) * time.Second,
		RPCAddr:           *rpcAddr,
		ModeChangeHandler: func(totalIterations uint64, modeNumber int) {},
		REPL:              true,
	}
	if err = ex.Execute(); err != nil {
		panic(err)
	}

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
				detector.NewDeviantDetector(positions),
			},
			ModeChangeHandler:  recordModeChanges,
			MissionFailurePlan: nextFailurePlan,
			OutputLocation:     *outputLocation,
		}
		doneChan := make(chan int)
		go func() {
			if err := ex.Execute(); err != nil {
				log.Fatalf("Error executing: %s\n", err)
			}
			doneChan <- 1
		}()

		select {
		case <-signals:
			log.Println("Received signal, exiting.")
			displayStats()
			os.Exit(0)
		case <-doneChan:
			// do nothing, this is normal
		}

		// update our statistics
		if !ex.MissionSuccessful {
			updateStats(nextFailurePlan)
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
			failurePlans = append(failurePlans, shiftedFailures)
		} // otherwise, we don't need to consider the shifted scenario

		enqueueScenarios(modeChangeTimes, &failurePlans, consideredScenarios)
		if err := saveModes(modeChangeTimes); err != nil {
			log.Printf("error saving mode transitions: %s", err)
		}
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
	sensorTypes := []hinj.Sensor{hinj.GPS, hinj.Accelerometer, hinj.Compass, hinj.Gyroscope, hinj.Barometer}
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
		results = append(results, copyAndAppend(set, failures[0]))
	}

	return results
}

// copies and appends
func copyAndAppend(failures []executor.FailurePlan, item executor.FailurePlan) []executor.FailurePlan {
	cp := make([]executor.FailurePlan, len(failures))
	copy(cp, failures)
	cp = append(cp, item)
	return cp
}

// called when a failure is encountered to record relevant statistics
func updateStats(failurePlan []executor.FailurePlan) {
	hasGPS, hasBaro, hasAccel, hasCompass, hasGyro := false, false, false, false, false
	for _, plan := range failurePlan {
		switch plan.SensorFailure.SensorType {
		case hinj.GPS:
			hasGPS = true
		case hinj.Barometer:
			hasBaro = true
		case hinj.Accelerometer:
			hasAccel = true
		case hinj.Compass:
			hasCompass = true
		case hinj.Gyroscope:
			hasGyro = true
		}
	}
	if hasGPS {
		statistics.unsafeFromGPS++
	}
	if hasBaro {
		statistics.unsafeFromAccel++
	}
	if hasCompass {
		statistics.unsafeFromCompass++
	}
	if hasAccel {
		statistics.unsafeFromAccel++
	}
	if hasGyro {
		statistics.unsafeFromGyro++
	}
	statistics.totalUnsafe++
}

func displayStats() {
	fmt.Println("Stats:")
	fmt.Printf("    %d total unsafe scenarios\n", statistics.totalUnsafe)
	fmt.Printf("    %d unsafe scenarios w/ a GPS fault\n", statistics.unsafeFromGPS)
	fmt.Printf("    %d unsafe scenarios w/ a Baro fault\n", statistics.unsafeFromBaros)
	fmt.Printf("    %d unsafe scenarios w/ a Accel fault\n", statistics.unsafeFromAccel)
	fmt.Printf("    %d unsafe scenarios w/ a Compass fault\n", statistics.unsafeFromCompass)
	fmt.Printf("    %d unsafe scenarios w/ a Gyro fault\n", statistics.unsafeFromGyro)
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

func getOutputLocation() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return path.Join(wd, "bugs/")
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

func getSensorOutputLocation(sensorName string) string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return path.Join(cwd, fmt.Sprintf("data/%s.json", sensorName))
}

func saveModes(modeChangeTimes []uint64) error {
	file, err := os.Create(*modeOutputDirectory)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.Encode(modeChangeTimes)
	return nil
}
