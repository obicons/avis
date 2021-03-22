package hinj

import (
	"fmt"
	"io"

	"github.com/obicons/avis/util"
)

type HINJReader struct {
	reader io.Reader
}

func (h *HINJReader) ReadMessage() (interface{}, error) {
	sensorType, err := h.readMessageType()
	if err != nil {
		return nil, err
	}

	msgSize, err := h.readMessageSize()
	if err != nil {
		return nil, err
	}

	msgBytes := make([]byte, msgSize)
	count, err := h.reader.Read(msgBytes)
	if err != nil {
		return nil, err
	} else if count != int(msgSize-msgPreambleSize) {
		return nil, fmt.Errorf("ReadMessage(): bad Read() of length %d for type %d", count, sensorType)
	}

	switch sensorType {
	case GPS:
		gpsPacket := GPSPacket{}
		if err := util.ReadPackedStruct(msgBytes, &gpsPacket); err != nil {
			return nil, fmt.Errorf("ReadMessage(): reading GPS: %s\n", err)
		}
		return &gpsPacket, nil
	case Accelerometer:
		accelerometerPacket := AccelerometerPacket{}
		if err := util.ReadPackedStruct(msgBytes, &accelerometerPacket); err != nil {
			return nil, fmt.Errorf("ReadMessage(): reading accelerometer: %s\n", err)
		}
		return &accelerometerPacket, nil
	case Gyroscope:
		gyroPacket := GyroscopePacket{}
		if err := util.ReadPackedStruct(msgBytes, &gyroPacket); err != nil {
			return nil, fmt.Errorf("ReadMessage(): reading gyro: %s\n", err)
		}
		return &gyroPacket, nil
	case Battery:
		batteryPacket := BatteryPacket{}
		if err := util.ReadPackedStruct(msgBytes, &batteryPacket); err != nil {
			return nil, fmt.Errorf("ReadMessage(): reading battery: %s\n", err)
		}
		return &batteryPacket, nil
	case Barometer:
		barometerPacket := BarometerPacket{}
		if err := util.ReadPackedStruct(msgBytes, &barometerPacket); err != nil {
			return nil, fmt.Errorf("ReadMessage(): reading barometer: %s\n", err)
		}
		return &barometerPacket, nil
	case Mode:
		modePacket := ModePacket{}
		if err := util.ReadPackedStruct(msgBytes, &modePacket); err != nil {
			return nil, fmt.Errorf("ReadMessage(): reading mode: %s\n", err)
		}
		return &modePacket, nil
	case Compass:
		compassPacket := CompassPacket{}
		if err := util.ReadPackedStruct(msgBytes, &compassPacket); err != nil {
			return nil, fmt.Errorf("ReadMessage(): reading compass: %s\n", err)
		}
		return &compassPacket, nil
	default:
		return nil, fmt.Errorf("ReadMessage(): unsupported type: %d", sensorType)
	}
}

func (h *HINJReader) readMessageType() (Sensor, error) {
	var typeByte [1]byte
	count, err := h.reader.Read(typeByte[:])
	if err != nil {
		return BadType, fmt.Errorf("readMessageType(): %s", err)
	} else if count != len(typeByte) {
		return BadType, fmt.Errorf("readMessageType(): could not read a complete type")
	} else if typeByte[0] > uint8(Mode) {
		return BadType, fmt.Errorf("readMessageType(): unknown type")
	}
	return Sensor(typeByte[0]), nil
}

func (h *HINJReader) readMessageSize() (uint32, error) {
	var sizeBytes [4]byte
	count, err := h.reader.Read(sizeBytes[:])
	if err != nil {
		return 0, fmt.Errorf("readMessageSize(): %s", err)
	} else if count != len(sizeBytes) {
		return 0, fmt.Errorf("readMessageSize(): could not read a complete size")
	}
	return util.HostByteOrder.Uint32(sizeBytes[:]), nil
}

func NewHINJReader(reader io.Reader) *HINJReader {
	hinjReader := new(HINJReader)
	hinjReader.reader = reader
	return hinjReader
}
