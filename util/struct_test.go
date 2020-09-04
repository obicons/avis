package util

import (
	"math"
	"testing"
)

type emptyStruct struct{}

type simpleStruct struct {
	U16 uint16
	I16 int16
	U32 uint32
	I32 int32
	U64 uint64
	I64 int64
	F32 float32
	F64 float64
}

type f32Struct struct {
	a float32
	b float32
	c float32
}

var (
	expectedU16 = uint16(42)
	expectedI16 = int16(-42)
	expectedU32 = uint32(42)
	expectedI32 = int32(-42)
	expectedU64 = uint64(42)
	expectedI64 = int64(-42)
	expectedF32 = float32(3.14)
	expectedF64 = float64(2.718)
)

func TestUnitReadPackedStructEmpty(t *testing.T) {
	var bytes []byte
	theStruct := emptyStruct{}
	if err := ReadPackedStruct(bytes, &theStruct); err != nil {
		t.Fatalf("error: ReadPackedStruct() returned an unexpected error: %s", err)
	}
}

func TestUnitReadPackedStructSimple(t *testing.T) {
	var bytes [40]byte
	var theStruct simpleStruct

	HostByteOrder.PutUint16(bytes[0:2], expectedU16)
	HostByteOrder.PutUint16(bytes[2:4], uint16(expectedI16))
	HostByteOrder.PutUint32(bytes[4:8], expectedU32)
	HostByteOrder.PutUint32(bytes[8:12], uint32(expectedI32))
	HostByteOrder.PutUint64(bytes[12:20], expectedU64)
	HostByteOrder.PutUint64(bytes[20:28], uint64(expectedI64))
	HostByteOrder.PutUint32(bytes[28:32], math.Float32bits(expectedF32))
	HostByteOrder.PutUint64(bytes[32:40], math.Float64bits(expectedF64))
	ReadPackedStruct(bytes[:], &theStruct)

	if theStruct.U16 != expectedU16 {
		t.Fatalf("Expected U16 = %d, found %d", expectedU16, theStruct.U16)
	}
	if theStruct.I16 != expectedI16 {
		t.Fatalf("Expected I16 = %d, found %d", expectedI16, theStruct.I16)
	}
	if theStruct.U32 != expectedU32 {
		t.Fatalf("Expected U32 = %d, found %d", expectedU32, theStruct.U32)
	}
	if theStruct.I32 != expectedI32 {
		t.Fatalf("Expected I32 = %d, found %d", expectedI32, theStruct.I32)
	}
	if theStruct.U64 != expectedU64 {
		t.Fatalf("Expected U64 = %d, found %d", expectedU64, theStruct.U64)
	}
	if theStruct.I64 != expectedI64 {
		t.Fatalf("Expected I64 = %d, found %d", expectedI64, theStruct.I64)
	}
	if theStruct.F32 != expectedF32 {
		t.Fatalf("Expected F32 = %f, found %f", expectedF32, theStruct.F32)
	}
	if theStruct.F64 != expectedF64 {
		t.Fatalf("Expected F64 = %f, found %f", expectedF64, theStruct.F64)
	}
}

func TestUnitPackedStructSizeEmptyValue(t *testing.T) {
	expectedSize := 0
	actualSize, err := PackedStructSize(emptyStruct{})
	if err != nil {
		t.Fatalf("PackedStructSize() returned an unexpected error: %s", err)
	}

	if expectedSize != actualSize {
		t.Fatalf("expected size = %d, found %d", expectedSize, actualSize)
	}
}

func TestUnitPackedStructSizeEmptyPtr(t *testing.T) {
	theStruct := emptyStruct{}
	expectedSize := 0
	actualSize, err := PackedStructSize(&theStruct)
	if err != nil {
		t.Fatalf("PackedStructSize() returned an unexpected error: %s", err)
	}

	if expectedSize != actualSize {
		t.Fatalf("expected size = %d, found %d", expectedSize, actualSize)
	}
}

func TestUnitPackedStructSizeSimpleValue(t *testing.T) {
	expectedSize := 40
	actualSize, err := PackedStructSize(simpleStruct{})
	if err != nil {
		t.Fatalf("PackedStructSize() returned an unexpected error: %s", err)
	}

	if expectedSize != actualSize {
		t.Fatalf("expected size = %d, found %d", expectedSize, actualSize)
	}
}

func TestUnitPackedStructSizeSimplePtr(t *testing.T) {
	theStruct := simpleStruct{}
	expectedSize := 40
	actualSize, err := PackedStructSize(&theStruct)
	if err != nil {
		t.Fatalf("PackedStructSize() returned an unexpected error: %s", err)
	}

	if expectedSize != actualSize {
		t.Fatalf("expected size = %d, found %d", expectedSize, actualSize)
	}
}

func TestUnitWritePackedStructEmptyValue(t *testing.T) {
	var buffer []byte
	theStruct := emptyStruct{}
	if err := PackedStructToBytes(buffer, theStruct); err != nil {
		t.Fatalf("PackedStructToBytes() returned an unexpected error: %s", err)
	}
}

func TestUnitWritePackedStructEmptyPtr(t *testing.T) {
	var buffer []byte
	theStruct := emptyStruct{}
	if err := PackedStructToBytes(buffer, &theStruct); err != nil {
		t.Fatalf("PackedStructToBytes() returned an unexpected error: %s", err)
	}
}

func TestUnitWritePackedStructSimpleValue(t *testing.T) {
	var buffer [40]byte
	theStruct := simpleStruct{
		U16: expectedU16,
		I16: expectedI16,
		U32: expectedU32,
		I32: expectedI32,
		U64: expectedU64,
		I64: expectedI64,
		F32: expectedF32,
		F64: expectedF64,
	}
	if err := PackedStructToBytes(buffer[:], theStruct); err != nil {
		t.Fatalf("PackedStructToBytes() returned an unexpected error: %s", err)
	}

	actualU16 := HostByteOrder.Uint16(buffer[:2])
	actualI16 := int16(HostByteOrder.Uint16(buffer[2:4]))
	actualU32 := HostByteOrder.Uint32(buffer[4:8])
	actualI32 := int32(HostByteOrder.Uint32(buffer[8:12]))
	actualU64 := HostByteOrder.Uint64(buffer[12:20])
	actualI64 := int64(HostByteOrder.Uint64(buffer[20:28]))
	actualF32 := math.Float32frombits(HostByteOrder.Uint32(buffer[28:32]))
	actualF64 := math.Float64frombits(HostByteOrder.Uint64(buffer[32:40]))

	if expectedU16 != actualU16 {
		t.Fatalf("Expected U16 = %d, found %d", expectedU16, actualU16)
	}
	if expectedI16 != actualI16 {
		t.Fatalf("Expected I16 = %d, found %d", expectedI16, actualI16)
	}
	if actualU32 != expectedU32 {
		t.Fatalf("Expected U32 = %d, found %d", expectedU32, actualU32)
	}
	if actualI32 != expectedI32 {
		t.Fatalf("Expected I32 = %d, found %d", expectedI32, actualI32)
	}
	if actualU64 != expectedU64 {
		t.Fatalf("Expected U64 = %d, found %d", expectedU64, actualU64)
	}
	if actualI64 != expectedI64 {
		t.Fatalf("Expected I64 = %d, found %d", expectedI64, actualI64)
	}
	if actualF32 != expectedF32 {
		t.Fatalf("Expected F32 = %f, found %f", expectedF32, actualF32)
	}
	if actualF64 != expectedF64 {
		t.Fatalf("Expected F64 = %f, found %f", expectedF64, actualF64)
	}
}

func TestUnitF32Size(t *testing.T) {
	expected := 12
	actual, _ := PackedStructSize(f32Struct{})
	if expected != actual {
		t.Fatalf("expected size %d, found %d", expected, actual)
	}
}
