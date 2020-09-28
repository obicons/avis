package detector

import (
	"math/rand"

	"github.com/obicons/rmck/entities"
)

type PositionRecorder struct {
	positionChan chan entities.TimestampedPosition
	shutdownChan chan int
	positions    []entities.Position
	generator    *rand.Rand
}

func NewPositionRecorder() Detector {
	return &PositionRecorder{
		positionChan: make(chan entities.TimestampedPosition),
		shutdownChan: make(chan int),
		generator:    rand.New(rand.NewSource(0)),
	}
}

func (pr *PositionRecorder) Start() {
	go func() {
		keepGoing := true
		for keepGoing {
			select {
			case <-pr.shutdownChan:
				keepGoing = false
			case pos := <-pr.positionChan:
				pr.samplePosition(pos.Position)
			}
		}
	}()
}

func (pr *PositionRecorder) Shutdown() {
	pr.shutdownChan <- 1
}

func (pr *PositionRecorder) PositionChan() chan<- entities.TimestampedPosition {
	return pr.positionChan
}

func (pr *PositionRecorder) SetAnomalyChan(ch chan<- Anomaly) {
	// no op
}

// stores the position with a .01% probability
func (pr *PositionRecorder) samplePosition(pos entities.Position) {
	if pr.generator.Float64() >= .99 {
		pr.positions = append(pr.positions, pos)
	}
}

func (pr *PositionRecorder) GetPositions() []entities.Position {
	return pr.positions
}
