package detector

import (
	"math"
	"time"

	"github.com/obicons/avis/entities"
)

type FreeFallDetector struct {
	anomalyChan    chan<- Anomaly
	shutdownChan   chan int
	positionChan   chan entities.TimestampedPosition
	lastUpdateTime time.Time
	lastPosition   entities.Position
}

const FreeFallThreshold = 9.8

func NewFreeFallDetector() Detector {
	return &FreeFallDetector{
		shutdownChan: make(chan int),
		positionChan: make(chan entities.TimestampedPosition),
	}
}

// implements Detector
func (d *FreeFallDetector) PositionChan() chan<- entities.TimestampedPosition {
	return d.positionChan
}

// implements Detector
func (d *FreeFallDetector) SetAnomalyChan(ch chan<- Anomaly) {
	d.anomalyChan = ch
}

// implements Detector
func (d *FreeFallDetector) Start() {
	go d.work()
}

// implements Detector
func (d *FreeFallDetector) Shutdown() {
	d.shutdownChan <- 0
}

func (d *FreeFallDetector) work() {
	keepGoing := true
	yVelocity := 0.0
	count := 0
	for keepGoing {
		select {
		case <-d.shutdownChan:
			keepGoing = false
		case pos := <-d.positionChan:
			timeDiff := float64(pos.Time.Sub(d.lastUpdateTime)) / float64(time.Second)
			newYVelocity := (pos.Position.Y - d.lastPosition.Y) / timeDiff
			accelY := math.Abs(newYVelocity-yVelocity) / timeDiff
			yVelocity = newYVelocity
			d.lastPosition = pos.Position
			d.lastUpdateTime = pos.Time
			if accelY > FreeFallThreshold {
				count++
			}
			if count > 10 {
				d.anomalyChan <- Anomaly{
					Time: pos.Time,
					Kind: FreeFall,
				}
			}
		}
	}
}
