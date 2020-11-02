package entities

import "time"

type Position struct {
	X float64
	Y float64
	Z float64
}

type TimestampedPosition struct {
	Position Position
	Time     time.Time
}

type SensorTraceParameters struct {
	TraceSensors         bool
	AccelTraceOutput     string
	GPSTraceOutput       string
	GyroTraceOutput      string
	CompassTraceOutput   string
	BarometerTraceOutput string
	PositionTraceOutput  string
}
