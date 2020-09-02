package hinj

import (
	"bytes"
	"testing"

	"github.com/obicons/rmck/util"
)

func TestUnitReadMessageTypeGPS(t *testing.T) {
	msgType := []byte{0} // GPS
	expectedType := GPS
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeSensor(t *testing.T) {
	msgType := []byte{1} // sensor reading
	expectedType := SensorReading
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeRCInput(t *testing.T) {
	msgType := []byte{2} // RC inputs
	expectedType := RCInputs
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeQuaternion(t *testing.T) {
	msgType := []byte{3} // Quaternion
	expectedType := Quaternion
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeAccel(t *testing.T) {
	msgType := []byte{4} // Accelerometer
	expectedType := Accelerometer
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeGyro(t *testing.T) {
	msgType := []byte{5} // Gyroscope
	expectedType := Gyroscope
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeBattery(t *testing.T) {
	msgType := []byte{6} // Battery
	expectedType := Battery
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeCompass(t *testing.T) {
	msgType := []byte{7} // Compass
	expectedType := Compass
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeBarometer(t *testing.T) {
	msgType := []byte{8} // Barometer
	expectedType := Barometer
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeMode(t *testing.T) {
	msgType := []byte{9} // Mode
	expectedType := Mode
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err != nil {
		t.Fatalf("readMessageType() returned an unexpected error: %s", err)
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageTypeUnknown(t *testing.T) {
	msgType := []byte{255} // Unknown type
	expectedType := BadType
	buffer := bytes.NewBuffer(msgType)
	reader := NewHINJReader(buffer)
	actualType, err := reader.readMessageType()
	if err == nil {
		t.Fatalf("readMessageType() should have returned an error")
	} else if actualType != expectedType {
		t.Fatalf("Expected type %d, found %d", expectedType, actualType)
	}
}

func TestUnitReadMessageSizeNormal(t *testing.T) {
	var msgSizeBytes [4]byte
	expectedMsgSize := uint32(16)
	util.HostByteOrder.PutUint32(msgSizeBytes[:], expectedMsgSize)
	buffer := bytes.NewBuffer(msgSizeBytes[:])
	hinjReader := NewHINJReader(buffer)
	actualMsgSize, err := hinjReader.readMessageSize()
	if err != nil {
		t.Fatalf("readMessageSize() returned an unexpected error: %s", err)
	} else if actualMsgSize != expectedMsgSize {
		t.Fatalf("readMessageSize() returned %d, expected %d", actualMsgSize, expectedMsgSize)
	}
}

func TestUnitReadMessageSizeError(t *testing.T) {
	var msgSizeBytes [4]byte
	expectedMsgSize := uint32(16)
	util.HostByteOrder.PutUint32(msgSizeBytes[:], expectedMsgSize)
	buffer := bytes.NewBuffer(msgSizeBytes[0:2])
	hinjReader := NewHINJReader(buffer)
	_, err := hinjReader.readMessageSize()
	if err == nil {
		t.Fatalf("readMessageSize() should return an error")
	}
}
