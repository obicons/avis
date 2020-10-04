package detector

import (
	"time"

	"github.com/obicons/rmck/entities"
)

type timeoutDetector struct {
	positionChan chan entities.TimestampedPosition
	anomalyChan  chan<- Anomaly
	shutdownChan chan int
	timeout      time.Duration
}

func NewTimeoutDetector(timeout time.Duration) Detector {
	return &timeoutDetector{
		positionChan: make(chan entities.TimestampedPosition),
		shutdownChan: make(chan int),
		timeout:      timeout,
	}
}

func (t *timeoutDetector) PositionChan() chan<- entities.TimestampedPosition {
	return t.positionChan
}

func (t *timeoutDetector) Start() {
	go func() {
		keepGoing := true
		timer := time.NewTimer(t.timeout)
		for keepGoing {
			select {
			case <-t.shutdownChan:
				timer.Stop()
				keepGoing = false
			case <-timer.C:
				t.anomalyChan <- Anomaly{Kind: Timeout}
			case <-t.positionChan:
				// do nothing
			}
		}
	}()
}

func (t *timeoutDetector) SetAnomalyChan(ch chan<- Anomaly) {
	t.anomalyChan = ch
}

func (t *timeoutDetector) Shutdown() {
	t.shutdownChan <- 0
}
