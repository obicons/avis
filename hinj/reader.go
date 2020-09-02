package hinj

import (
	"fmt"
	"io"
	"log"

	"github.com/obicons/rmck/util"
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
	} else if count != int(msgSize) {
		return nil, fmt.Errorf("ReadMessage(): bad Read() of length %d", count)
	}

	switch sensorType {
	case GPS:
		gpsPacket := GPSPacket{}
		if err := util.ReadPackedStruct(msgBytes, gpsPacket); err != nil {
			log.Printf("ReadMessage(): reading GPS: %s\n", err)
		}
		return &gpsPacket, nil
		// TODO: complete me!
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
