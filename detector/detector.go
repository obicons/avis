package detector

import (
	"time"

	"github.com/obicons/rmck/entities"
)

type AnomalyKind int

const (
	AnomalyUnkown AnomalyKind = iota
	FreeFall
	ProgramFault
	Timeout
)

type Anomaly struct {
	// time of occurence
	Time time.Time

	// kind of anomaly
	Kind AnomalyKind
}

type Detector interface {
	// returns a channel to write new positions to
	PositionChan() chan<- entities.TimestampedPosition
	// stops detection immediately
	Shutdown()
	// starts detection
	Start()
	// sets the channel where anomalies are reported
	SetAnomalyChan(chan<- Anomaly)
}

func (k AnomalyKind) String() string {
	switch k {
	case AnomalyUnkown:
		return "Unknown anomaly"
	case FreeFall:
		return "Free Fall"
	case ProgramFault:
		return "Program Fault"
	case Timeout:
		return "Timeout"
	}
	return "Unknown anomaly"
}

func (a Anomaly) String() string {
	return a.Kind.String() + "@ " + a.Time.String()
}
