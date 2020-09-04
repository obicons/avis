package hinj

type Sensor uint8

const (
	GPS           Sensor = iota
	SensorReading        // currently unused
	RCInputs             // currently unused
	Quaternion           // currently unused
	Accelerometer
	Gyroscope
	Battery
	Compass
	Barometer
	Mode
	BadType
)

const (
	msgPreambleSize = 5
)

type SensorFailure struct {
	SensorType Sensor
	Instance   uint8
}

type GPSPacket struct {
	Instance          uint8
	Ignore            uint8
	TimeMicroSecond   uint64
	FixType           uint8
	Latitude          int32
	Longitude         int32
	Altitude          int32
	EPH               uint16
	EPV               uint16
	Velocity          uint16
	VelocityNorth     int16
	VelocityEast      int16
	VelocityDown      int16
	CourseOverGround  uint16
	SatellitesVisible uint8
}

type AccelerometerPacket struct {
	Instance      uint8
	Ignore        uint8
	AccelerationX float32
	AccelerationY float32
	AccelerationZ float32
}

type GyroscopePacket struct {
	Instance uint8
	Ignore   uint8
	X        float32
	Y        float32
	Z        float32
}

type BatteryPacket struct {
	Voltage  float32
	Current  float32
	Throttle float32
}

type BarometerPacket struct {
	Instance    uint8
	Ignore      uint8
	Pressure    float32
	Temperature float32
}

type CompassPacket struct {
	Instance uint8
	Ignore   uint8
	Mag0     float32
	Mag1     float32
	Mag2     float32
}

type ModePacket struct {
	Mode uint32
}
