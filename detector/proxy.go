package detector

import (
	"github.com/obicons/rmck/entities"
)

type DetectorProxy struct {
	positionChan    chan entities.TimestampedPosition
	anomalyChan     chan<- Anomaly
	detectors       []Detector
	shutdownChan    chan int
	shutdownAckChan chan int
}

func NewDetectorProxy(detectors []Detector, anomalyCh chan<- Anomaly) *DetectorProxy {
	return &DetectorProxy{
		positionChan:    make(chan entities.TimestampedPosition),
		shutdownChan:    make(chan int),
		shutdownAckChan: make(chan int),
		detectors:       detectors,
		anomalyChan:     anomalyCh,
	}
}

func (d *DetectorProxy) RegisterDetector(detector Detector) {
	d.detectors = append(d.detectors, detector)
}

func (d *DetectorProxy) Start() {
	// start our detectors
	for _, detector := range d.detectors {
		detector.SetAnomalyChan(d.anomalyChan)
		detector.Start()
	}

	// forward data to each detector
	go func() {
		shutdown := d.shutdownChan
		position := d.positionChan
		keepGoing := true
		for keepGoing {
			select {
			case <-shutdown:
				d.stopAllDetectors()
				keepGoing = false
			case pos := <-position:
				d.forwardPosition(pos)
			}
		}
		d.shutdownAckChan <- 1
	}()
}

func (d *DetectorProxy) Shutdown() {
	d.shutdownChan <- 0
	<-d.shutdownAckChan
}

func (d *DetectorProxy) PositionChan() chan entities.TimestampedPosition {
	return d.positionChan
}

func (d *DetectorProxy) SetAnomalyChan(ch chan<- Anomaly) {
	d.anomalyChan = ch
}

func (d *DetectorProxy) stopAllDetectors() {
	for _, detector := range d.detectors {
		detector.Shutdown()
	}
}

func (d *DetectorProxy) forwardPosition(p entities.TimestampedPosition) {
	for _, detector := range d.detectors {
		detector.PositionChan() <- p
	}
}
