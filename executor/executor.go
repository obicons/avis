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

	"github.com/obicons/rmck/controller"
	"github.com/obicons/rmck/detector"
	"github.com/obicons/rmck/entities"
	"github.com/obicons/rmck/hinj"
	"github.com/obicons/rmck/platforms"
	"github.com/obicons/rmck/sim"
	"github.com/obicons/rmck/util"
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
	rpcServer          *controller.SimulatorController
	MissionSuccessful  bool
	TraceParameters    entities.SensorTraceParameters
	accelPackets       map[uint64]hinj.AccelerometerPacket
	gyroPackets        map[uint64]hinj.GyroscopePacket
	gpsPackets         map[uint64]hinj.GPSPacket
	baroPackets        map[uint64]hinj.BarometerPacket
	compassPackets     map[uint64]hinj.CompassPacket
	positionData       map[uint64]entities.Position
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

	cmd := executeWorkload(e.WorkloadCmd)
	defer func() {
		cmd.Process.Kill()
		cmd.Process.Wait()
	}()

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

	dataSets := []interface{}{
		e.accelPackets,
		e.gpsPackets,
		e.gyroPackets,
		e.baroPackets,
		e.compassPackets,
		e.positionData,
	}
	dataSetFiles := []string{
		e.TraceParameters.AccelTraceOutput,
		e.TraceParameters.GPSTraceOutput,
		e.TraceParameters.GyroTraceOutput,
		e.TraceParameters.BarometerTraceOutput,
		e.TraceParameters.CompassTraceOutput,
		e.TraceParameters.PositionTraceOutput,
	}

	for idx, dataSet := range dataSets {
		outputLocation := dataSetFiles[idx]
		file, err := os.Create(outputLocation)
		if err != nil {
			log.Printf("error creating file: %s\n", err)
			continue
		}
		encoder := json.NewEncoder(file)
		encoder.Encode(dataSet)
		file.Close()
	}
}

func (e *Executor) clearSensors() {
	e.accelPackets = make(map[uint64]hinj.AccelerometerPacket)
	e.gyroPackets = make(map[uint64]hinj.GyroscopePacket)
	e.baroPackets = make(map[uint64]hinj.BarometerPacket)
	e.gpsPackets = make(map[uint64]hinj.GPSPacket)
	e.compassPackets = make(map[uint64]hinj.CompassPacket)
	e.positionData = make(map[uint64]entities.Position)
}

func (e *Executor) sampleSensors() {
	if e.rand.Float64() > .99 {
		it := e.Simulator.Iterations()
		e.accelPackets[it] = e.HINJServer.GetLastAccelReading()
		e.baroPackets[it] = e.HINJServer.GetLastBarometerReading()
		e.gpsPackets[it] = e.HINJServer.GetLastGPSReading()
		e.gyroPackets[it] = e.HINJServer.GetLastGyroReading()
		e.compassPackets[it] = e.HINJServer.GetLastCompassReading()
		pos, err := e.Simulator.Position(context.Background())
		if err != nil {
			log.Printf("sampleSensors(): error sampling position: %s\n", err)
		} else {
			e.positionData[it] = pos
		}
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
