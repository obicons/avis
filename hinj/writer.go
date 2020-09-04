package hinj

import (
	"fmt"
	"io"

	"github.com/obicons/rmck/util"
)

type HINJWriter struct {
	writer io.Writer
}

func (h *HINJWriter) WriteMessage(msg interface{}) error {
	sensor := BadType
	ok := false
	switch msg.(type) {
	case *GPSPacket:
		ok = true
		sensor = GPS
	case *AccelerometerPacket:
		ok = true
		sensor = Accelerometer
	case *GyroscopePacket:
		ok = true
		sensor = Gyroscope
	case *BatteryPacket:
		ok = true
		sensor = Battery
	case *ModePacket:
		ok = true
		sensor = Mode
	case *BarometerPacket:
		ok = true
		sensor = Barometer
	case *CompassPacket:
		ok = true
		sensor = Compass
	}

	if !ok {
		return fmt.Errorf("error: WriteMessage(): unrecognized type")
	}

	// the way we designed our packets implies there is never an error
	size, _ := util.PackedStructSize(msg)

	bytes := make([]byte, size+msgPreambleSize)

	// writes the preamble
	bytes[0] = byte(sensor)
	util.HostByteOrder.PutUint32(bytes[1:5], uint32(size+msgPreambleSize))

	// writes the rest of the packet
	util.PackedStructToBytes(bytes[msgPreambleSize:], msg)

	h.writer.Write(bytes)

	return nil
}

func NewHINJWriter(writer io.Writer) *HINJWriter {
	hw := HINJWriter{writer: writer}
	return &hw
}
