package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/obicons/avis/controller"
	"github.com/obicons/avis/detector"
	"github.com/obicons/avis/entities"
	"github.com/obicons/avis/hinj"
	"github.com/obicons/avis/platforms"
	"github.com/obicons/avis/sim"
	"github.com/obicons/avis/util"
)

type FailurePlan struct {
	SensorFailure hinj.SensorFailure
	// measured in iterations
	FailureTime uint64
}

type Executor struct {
	HINJServer         *hinj.HINJServer
	Simulator          sim.Sim
	Autopilot          platforms.System
	WorkloadCmd        string
	Timeout            time.Duration
	RPCAddr            string
	Detectors          []detector.Detector
	ModeChangeHandler  func(totalIterations uint64, modeNumber int)
	MissionFailurePlan []FailurePlan
	OutputLocation     string
	MissionSuccessful  bool
	TraceParameters    entities.SensorTraceParameters
	REPL               bool
	rpcServer          *controller.SimulatorController
	accelPackets       map[uint64]hinj.AccelerometerPacket
	gyroPackets        map[uint64]hinj.GyroscopePacket
	gpsPackets         map[uint64]hinj.GPSPacket
	baroPackets        map[uint64]hinj.BarometerPacket
	compassPackets     map[uint64]hinj.CompassPacket
	rand               *rand.Rand
}

func (e *Executor) Execute() error {
	e.clearSensors()
	e.rand = rand.New(rand.NewSource(42))

	var err error
	if err := e.HINJServer.Start(); err != nil {
		return err
	}
	defer e.HINJServer.Shutdown()

	if err := e.Simulator.Start(); err != nil {
		return err
	}
	defer e.Simulator.Shutdown(context.Background())

	time.Sleep(time.Second * 5)

	if err := e.Autopilot.Start(); err != nil {
		return err
	}
	defer e.Autopilot.Shutdown(context.Background())

	e.rpcServer, err = controller.New(e.RPCAddr, e.Simulator)
	if err != nil {
		return err
	}

	go func() {
		// TODO -- handle error
		e.rpcServer.Start()
	}()
	defer e.rpcServer.Shutdown()

	// there needs to be appropriate space in this channel to avoid deadlock
	anomalyChan := make(chan detector.Anomaly, len(e.Detectors))
	detectorProxy := detector.NewDetectorProxy(e.Detectors, anomalyChan)
	modeExitChan := e.doModeReporting()
	detectorProxy.Start()
	defer func() { modeExitChan <- 1 }()
	defer detectorProxy.Shutdown()

	e.Simulator.AddPostStepAction(
		func() {
			ctx, cc := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cc()
			pos, err := e.Simulator.Position(ctx)
			if err != nil {
				return
			}

			ctx, cc = context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cc()
			time, err := e.Simulator.SimTime(ctx)
			if err != nil {
				return
			}

			if e.TraceParameters.TraceSensors {
				e.sampleSensors()
			}

			detectorProxy.PositionChan() <- entities.TimestampedPosition{
				Time:     time,
				Position: pos,
			}
		},
	)
	e.Simulator.AddPostStepAction(
		func() {
			// check if its time for a failure
			for _, plan := range e.MissionFailurePlan {
				if plan.FailureTime == e.Simulator.Iterations() {
					e.HINJServer.FailSensor(plan.SensorFailure.SensorType, plan.SensorFailure.Instance)
				}
			}
		},
	)

	time.Sleep(time.Second * 10)

	if !e.REPL {
		cmd := executeWorkload(e.WorkloadCmd)
		defer func() {
			cmd.Process.Kill()
			cmd.Process.Wait()
		}()
	}

	rpcDone := e.rpcServer.Done()
	keepGoing := true

	// TODO: this can deadlock if a client calls Terminate, but then reports mode changes.
	// It can also deadlock if the anomaly queue is saturated between a client terminating and shutting down detectors
	for keepGoing {
		select {
		case <-rpcDone:
			keepGoing = false
			e.MissionSuccessful = true
		case anomaly := <-anomalyChan:
			ts := time.Now()
			fmt.Printf("Anomaly detected: %s\n", anomaly.String())
			outputFilePath := path.Join(e.OutputLocation, strconv.FormatInt(ts.Unix(), 10))
			file, err := os.Create(outputFilePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error saving trace: %s\n", err)
			} else {
				encoder := json.NewEncoder(file)
				encoder.Encode(e.MissionFailurePlan)
				file.Close()
			}
			e.MissionSuccessful = false
			keepGoing = false
		}
	}

	e.maybeSaveSensors()

	return nil
}

func (e *Executor) maybeSaveSensors() {
	if !e.TraceParameters.TraceSensors {
		return
	}

	accelFile, err := os.Create(e.TraceParameters.AccelTraceOutput)
	if err != nil {
		log.Printf("unable to create accel trace: %s\n", err)
		return
	}
	accelEncoder := json.NewEncoder(accelFile)
	accelEncoder.Encode(e.accelPackets)
	accelFile.Close()

	gpsFile, err := os.Create(e.TraceParameters.GPSTraceOutput)
	if err != nil {
		log.Printf("unable to create gps trace: %s\n", err)
		return
	}
	gpsEncoder := json.NewEncoder(gpsFile)
	gpsEncoder.Encode(e.gpsPackets)
	gpsFile.Close()

	gyroFile, err := os.Create(e.TraceParameters.GyroTraceOutput)
	if err != nil {
		log.Printf("unable to create gyro trace: %s\n", err)
		return
	}
	gyroEncoder := json.NewEncoder(gyroFile)
	gyroEncoder.Encode(e.gyroPackets)
	gyroFile.Close()

	baroFile, err := os.Create(e.TraceParameters.BarometerTraceOutput)
	if err != nil {
		log.Printf("unable to create baro trace: %s\n", err)
		return
	}
	baroEncoder := json.NewEncoder(baroFile)
	baroEncoder.Encode(e.baroPackets)
	baroFile.Close()

	compassFile, err := os.Create(e.TraceParameters.CompassTraceOutput)
	if err != nil {
		log.Printf("unable to create compass trace: %s\n", err)
		return
	}
	compassEncoder := json.NewEncoder(compassFile)
	compassEncoder.Encode(e.compassPackets)
	compassFile.Close()

}

func (e *Executor) clearSensors() {
	e.accelPackets = make(map[uint64]hinj.AccelerometerPacket)
	e.gyroPackets = make(map[uint64]hinj.GyroscopePacket)
	e.baroPackets = make(map[uint64]hinj.BarometerPacket)
	e.gpsPackets = make(map[uint64]hinj.GPSPacket)
	e.compassPackets = make(map[uint64]hinj.CompassPacket)
}

func (e *Executor) sampleSensors() {
	if e.rand.Float64() > .99 {
		it := e.Simulator.Iterations()
		e.accelPackets[it] = e.HINJServer.GetLastAccelReading()
		e.baroPackets[it] = e.HINJServer.GetLastBarometerReading()
		e.gpsPackets[it] = e.HINJServer.GetLastGPSReading()
		e.gyroPackets[it] = e.HINJServer.GetLastGyroReading()
		e.compassPackets[it] = e.HINJServer.GetLastCompassReading()
	}
}

func (e *Executor) doModeReporting() chan int {
	keepGoing := true
	exitCh := make(chan int)
	go func() {
		modeCh := e.rpcServer.Mode()
		for keepGoing {
			select {
			case <-exitCh:
				keepGoing = false
			case mode := <-modeCh:
				if e.ModeChangeHandler != nil {
					iterations := e.Simulator.Iterations()
					e.ModeChangeHandler(iterations, mode)
				}
			}
		}
	}()
	return exitCh
}

func executeWorkload(workloadCmd string) *exec.Cmd {
	log, err := util.GetLogger("workload ")
	if err != nil {
		log.Fatalf("Could not get a logger: %s", err)
	}

	cmd := exec.Command("sh", "-c", workloadCmd)
	cmd.Env = os.Environ()
	util.LogProcess(cmd, log)
	cmd.Start()

	return cmd
}
