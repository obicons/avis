package detector

import (
	"math/rand"

	"github.com/obicons/rmck/entities"
	"github.com/obicons/rmck/util"
)

const threshold = 10
const seed = 42

type DeviantDetector struct {
	goldenRunData []entities.Position
	positionChan  chan entities.TimestampedPosition
	anomalyChan   chan<- Anomaly
	shutdownChan  chan int
	reported      bool
	rand          *rand.Rand
}

// returns a new instance of DeviantDetector
func NewDeviantDetector(goldenRunData []entities.Position) Detector {
	return &DeviantDetector{
		goldenRunData: goldenRunData,
		positionChan:  make(chan entities.TimestampedPosition),
		shutdownChan:  make(chan int),
		reported:      false,
		rand:          rand.New(rand.NewSource(seed)),
	}
}

// implements detector.Detector
func (d *DeviantDetector) PositionChan() chan<- entities.TimestampedPosition {
	return d.positionChan
}

// implements detector.Detector
func (d *DeviantDetector) SetAnomalyChan(ch chan<- Anomaly) {
	d.anomalyChan = ch
}

// implements detector.Detector
func (d *DeviantDetector) Start() {
	go func() {
		keepGoing := true
		i, count := 0, 0
		for keepGoing {
			select {
			case <-d.shutdownChan:
				keepGoing = false
			case pos := <-d.positionChan:
				// implements sampling
				if d.rand.Float64() <= .99 {
					continue
				}

				if i < len(d.goldenRunData) && !d.reported {
					dist := util.Distance(d.goldenRunData[i], pos.Position)
					if dist > threshold {
						count++
						if count > 50 {
							d.reported = true
							d.anomalyChan <- Anomaly{
								Time: pos.Time,
								Kind: Deviation,
							}
						}
					} else {
						count = 0
					}
				}
				i++
			}
		}
	}()
}

// implements detector.Detector
func (d *DeviantDetector) Shutdown() {
	d.shutdownChan <- 0
}
