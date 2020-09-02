package util

import (
	"fmt"
	"math"
	"reflect"
)

// Reads a packed struct.
// place must point to a struct with primitive members only
func ReadPackedStruct(bytes []byte, place interface{}) error {
	if place == nil {
		return fmt.Errorf("error: ReadPackedStruct(): place is nil")
	}

	val := reflect.Indirect(reflect.ValueOf(place))
	t := val.Type()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("error: ReadPackedStruct(): type %s is not a struct", t.Name())
	}

	for fieldNo := 0; fieldNo < t.NumField(); fieldNo++ {
		field := t.Field(fieldNo)
		switch field.Type.Kind() {
		case reflect.Uint8:
			if len(bytes) < 1 {
				return fmt.Errorf("error: not enough bytes to read uint8: %d", len(bytes))
			}
			u8 := uint8(HostByteOrder.Uint16(bytes[0:1]))
			val.Field(fieldNo).Set(reflect.ValueOf(u8))
			bytes = bytes[1:]
		case reflect.Int8:
			if len(bytes) < 1 {
				return fmt.Errorf("error: not enough bytes to read int8: %d", len(bytes))
			}
			i8 := int8(HostByteOrder.Uint16(bytes[0:1]))
			val.Field(fieldNo).Set(reflect.ValueOf(i8))
			bytes = bytes[1:]
		case reflect.Uint16:
			if len(bytes) < 2 {
				return fmt.Errorf("error: not enough bytes to read uint16: %d", len(bytes))
			}
			u16 := HostByteOrder.Uint16(bytes[0:2])
			val.Field(fieldNo).Set(reflect.ValueOf(u16))
			bytes = bytes[2:]
		case reflect.Int16:
			if len(bytes) < 2 {
				return fmt.Errorf("error: not enough bytes to read int16: %d", len(bytes))
			}
			i16 := int16(HostByteOrder.Uint16(bytes[0:2]))
			val.Field(fieldNo).Set(reflect.ValueOf(i16))
			bytes = bytes[2:]
		case reflect.Uint32:
			if len(bytes) < 4 {
				return fmt.Errorf("error: not enough bytes to read uint32: %d", len(bytes))
			}
			u32 := HostByteOrder.Uint32(bytes[0:4])
			val.Field(fieldNo).Set(reflect.ValueOf(u32))
			bytes = bytes[4:]
		case reflect.Int32:
			if len(bytes) < 4 {
				return fmt.Errorf("error: not enough bytes to read int32: %d", len(bytes))
			}
			i32 := int32(HostByteOrder.Uint32(bytes[0:4]))
			val.Field(fieldNo).Set(reflect.ValueOf(i32))
			bytes = bytes[4:]
		case reflect.Uint64:
			if len(bytes) < 8 {
				return fmt.Errorf("error: not enough bytes to read uint64: %d", len(bytes))
			}
			u64 := HostByteOrder.Uint64(bytes[0:8])
			val.Field(fieldNo).Set(reflect.ValueOf(u64))
			bytes = bytes[8:]
		case reflect.Int64:
			if len(bytes) < 8 {
				return fmt.Errorf("error: not enough bytes to read int64: %d", len(bytes))
			}
			i64 := int64(HostByteOrder.Uint64(bytes[0:8]))
			val.Field(fieldNo).Set(reflect.ValueOf(i64))
			bytes = bytes[8:]
		case reflect.Float32:
			if len(bytes) < 4 {
				return fmt.Errorf("error: not enough bytes to read float32: %d", len(bytes))
			}
			f32 := math.Float32frombits(HostByteOrder.Uint32(bytes[0:4]))
			val.Field(fieldNo).Set(reflect.ValueOf(f32))
			bytes = bytes[4:]
		case reflect.Float64:
			if len(bytes) < 8 {
				return fmt.Errorf("error: not enough bytes to read float64: %d", len(bytes))
			}
			f64 := math.Float64frombits(HostByteOrder.Uint64(bytes[0:8]))
			val.Field(fieldNo).Set(reflect.ValueOf(f64))
			bytes = bytes[8:]

		default:
			return fmt.Errorf("error: cannot read non-primitive type %s", field.Type.Name())
		}
	}

	return nil
}

// returns the size of the packed struct.
// It is an error to call this function on anything that is not a struct (or a pointer thereto) with primitive-only members.
func PackedStructSize(theStruct interface{}) (int, error) {
	size := 0

	if theStruct == nil {
		return size, nil
	}

	t := reflect.TypeOf(theStruct)
	v := reflect.ValueOf(theStruct)
	if t.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		return 0, fmt.Errorf("error: PackedStructSize() called on non-struct type %s", t.Name())
	}

	for fieldNo := 0; fieldNo < v.NumField(); fieldNo++ {
		field := t.Field(fieldNo)
		size += int(field.Type.Size())
	}

	return size, nil
}

// Converts theStruct into an array of bytes.
// It is an error to call this function on anything that is not a struct (or a pointer thereto) with primitive-only members.
// len(bytes) must be sufficient to store every member of theStruct.
func PackedStructToBytes(bytes []byte, theStruct interface{}) error {
	t := reflect.TypeOf(theStruct)
	v := reflect.ValueOf(theStruct)
	if t.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("error: PackedStructToBytes() called on non-struct type %s", t.Name())
	}

	for fieldNo := 0; fieldNo < v.NumField(); fieldNo++ {
		field := t.Field(fieldNo)
		switch field.Type.Kind() {
		case reflect.Uint8:
			if len(bytes) < 1 {
				return fmt.Errorf("error: not enough space to write uint8: %d", len(bytes))
			}
			bytes[0] = byte(v.Field(fieldNo).Uint())
			bytes = bytes[1:]
		case reflect.Int8:
			if len(bytes) < 1 {
				return fmt.Errorf("error: not enough space to write int8: %d", len(bytes))
			}
			bytes[0] = byte(v.Field(fieldNo).Int())
			bytes = bytes[1:]
		case reflect.Uint16:
			if len(bytes) < 2 {
				return fmt.Errorf("error: not enough space to write uint16: %d", len(bytes))
			}
			HostByteOrder.PutUint16(bytes[:2], uint16(v.Field(fieldNo).Uint()))
			bytes = bytes[2:]
		case reflect.Int16:
			if len(bytes) < 2 {
				return fmt.Errorf("error: not enough space to write int16: %d", len(bytes))
			}
			HostByteOrder.PutUint16(bytes[:2], uint16(v.Field(fieldNo).Int()))
			bytes = bytes[2:]
		case reflect.Uint32:
			if len(bytes) < 4 {
				return fmt.Errorf("error: not enough space to write uint32: %d", len(bytes))
			}
			HostByteOrder.PutUint32(bytes[:4], uint32(v.Field(fieldNo).Uint()))
			bytes = bytes[4:]
		case reflect.Int32:
			if len(bytes) < 4 {
				return fmt.Errorf("error: not enough space to write int32: %d", len(bytes))
			}
			HostByteOrder.PutUint32(bytes[:4], uint32(v.Field(fieldNo).Int()))
			bytes = bytes[4:]
		case reflect.Uint64:
			if len(bytes) < 8 {
				return fmt.Errorf("error: not enough space to write uint64: %d", len(bytes))
			}
			HostByteOrder.PutUint64(bytes[:8], v.Field(fieldNo).Uint())
			bytes = bytes[8:]
		case reflect.Int64:
			if len(bytes) < 8 {
				return fmt.Errorf("error: not enough space to write int64: %d", len(bytes))
			}
			HostByteOrder.PutUint64(bytes[:8], uint64(v.Field(fieldNo).Int()))
			bytes = bytes[8:]
		case reflect.Float32:
			if len(bytes) < 4 {
				return fmt.Errorf("error: not enough space to write float32: %d", len(bytes))
			}
			HostByteOrder.PutUint32(bytes[:4], math.Float32bits(float32(v.Field(fieldNo).Float())))
			bytes = bytes[4:]
		case reflect.Float64:
			if len(bytes) < 8 {
				return fmt.Errorf("error: not enough bytes to read float64: %d", len(bytes))
			}
			HostByteOrder.PutUint64(bytes[:8], math.Float64bits(v.Field(fieldNo).Float()))
			bytes = bytes[8:]

		default:
			return fmt.Errorf("error: cannot read non-primitive type %s", field.Type.Name())
		}
	}

	return nil
}
