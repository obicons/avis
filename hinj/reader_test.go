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

func TestUnitReadMessageGPS(t *testing.T) {
	gps := GPSPacket{Instance: 1, SatellitesVisible: 10}
	size, _ := util.PackedStructSize(&gps)
	msgBytes := make([]byte, msgPreambleSize+size)
	msgBytes[0] = byte(GPS)
	util.HostByteOrder.PutUint32(msgBytes[1:5], uint32(size+msgPreambleSize))
	util.PackedStructToBytes(msgBytes[5:], &gps)
	buffer := bytes.NewBuffer(msgBytes)
	reader := NewHINJReader(buffer)
	gpsInterface, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() returned an unexpected error: %s", err)
	}
	if realGPS, ok := gpsInterface.(*GPSPacket); !ok {
		t.Fatalf("ReadMessage() did not return the expected type")
	} else if realGPS.Instance != 1 || realGPS.SatellitesVisible != 10 {
		t.Fatalf("ReadMessage() returned a GPS packet with incorrect settings")
	}
}

func TestUnitReadMessageAccelerometer(t *testing.T) {
	accel := AccelerometerPacket{Instance: 1, AccelerationX: 10.0}
	size, _ := util.PackedStructSize(&accel)
	msgBytes := make([]byte, msgPreambleSize+size)
	msgBytes[0] = byte(Accelerometer)
	util.HostByteOrder.PutUint32(msgBytes[1:5], uint32(size+msgPreambleSize))
	util.PackedStructToBytes(msgBytes[5:], &accel)
	buffer := bytes.NewBuffer(msgBytes)
	reader := NewHINJReader(buffer)
	accelInterface, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() returned an unexpected error: %s", err)
	}
	if realAccel, ok := accelInterface.(*AccelerometerPacket); !ok {
		t.Fatalf("ReadMessage() did not return the expected type")
	} else if realAccel.Instance != 1 || realAccel.AccelerationX != 10.0 {
		t.Fatalf("ReadMessage() returned an Accelerometer packet with incorrect settings")
	}
}

func TestUnitReadMessageGyro(t *testing.T) {
	gyro := GyroscopePacket{Instance: 1, Z: 2.0}
	size, _ := util.PackedStructSize(&gyro)
	msgBytes := make([]byte, msgPreambleSize+size)
	msgBytes[0] = byte(Gyroscope)
	util.HostByteOrder.PutUint32(msgBytes[1:5], uint32(size+msgPreambleSize))
	util.PackedStructToBytes(msgBytes[5:], &gyro)
	buffer := bytes.NewBuffer(msgBytes)
	reader := NewHINJReader(buffer)
	gyroInterface, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() returned an unexpected error: %s", err)
	}
	if realGyro, ok := gyroInterface.(*GyroscopePacket); !ok {
		t.Fatalf("ReadMessage() did not return the expected type")
	} else if realGyro.Instance != 1 || realGyro.Z != 2.0 {
		t.Fatalf("ReadMessage() returned a Gyroscope packet with incorrect settings")
	}
}

func TestUnitReadMessageBattery(t *testing.T) {
	battery := BatteryPacket{Voltage: 1.0, Throttle: 2.0}
	size, _ := util.PackedStructSize(&battery)
	msgBytes := make([]byte, msgPreambleSize+size)
	msgBytes[0] = byte(Battery)
	util.HostByteOrder.PutUint32(msgBytes[1:5], uint32(size+msgPreambleSize))
	util.PackedStructToBytes(msgBytes[5:], &battery)
	buffer := bytes.NewBuffer(msgBytes)
	reader := NewHINJReader(buffer)
	batteryInterface, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() returned an unexpected error: %s", err)
	}
	if realBattery, ok := batteryInterface.(*BatteryPacket); !ok {
		t.Fatalf("ReadMessage() did not return the expected type")
	} else if realBattery.Voltage != 1.0 || realBattery.Throttle != 2.0 {
		t.Fatalf("ReadMessage() returned a Battery packet with incorrect settings")
	}
}

func TestUnitReadMessageBarometer(t *testing.T) {
	baro := BarometerPacket{Instance: 1, Temperature: 42.0}
	size, _ := util.PackedStructSize(&baro)
	msgBytes := make([]byte, msgPreambleSize+size)
	msgBytes[0] = byte(Barometer)
	util.HostByteOrder.PutUint32(msgBytes[1:5], uint32(size+msgPreambleSize))
	util.PackedStructToBytes(msgBytes[5:], &baro)
	buffer := bytes.NewBuffer(msgBytes)
	reader := NewHINJReader(buffer)
	baroInterface, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() returned an unexpected error: %s", err)
	}
	if realBarometer, ok := baroInterface.(*BarometerPacket); !ok {
		t.Fatalf("ReadMessage() did not return the expected type")
	} else if realBarometer.Instance != 1 || realBarometer.Temperature != 42.0 {
		t.Fatalf("ReadMessage() returned a Barometer packet with incorrect settings")
	}
}

func TestUnitReadMessageMode(t *testing.T) {
	mode := ModePacket{Mode: 42}
	size, _ := util.PackedStructSize(&mode)
	msgBytes := make([]byte, msgPreambleSize+size)
	msgBytes[0] = byte(Mode)
	util.HostByteOrder.PutUint32(msgBytes[1:5], uint32(size+msgPreambleSize))
	util.PackedStructToBytes(msgBytes[5:], &mode)
	buffer := bytes.NewBuffer(msgBytes)
	reader := NewHINJReader(buffer)
	modeInterface, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() returned an unexpected error: %s", err)
	}
	if realMode, ok := modeInterface.(*ModePacket); !ok {
		t.Fatalf("ReadMessage() did not return the expected type")
	} else if realMode.Mode != 42 {
		t.Fatalf("ReadMessage() returned a Mode packet with incorrect settings")
	}
}
