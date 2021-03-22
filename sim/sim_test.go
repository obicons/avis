package sim

import (
	"math"
	"testing"

	"github.com/obicons/avis/util"
)

func TestUnitReadPosition(t *testing.T) {
	actualX, actualY, actualZ := 10.0, 20.0, 40.0
	var bytes [24]byte
	position := Position{}
	util.HostByteOrder.PutUint64(bytes[0:8], math.Float64bits(actualX))
	util.HostByteOrder.PutUint64(bytes[8:16], math.Float64bits(actualY))
	util.HostByteOrder.PutUint64(bytes[16:24], math.Float64bits(actualZ))
	err := util.ReadPackedStruct(bytes[:], &position)
	if err != nil {
		t.Fatalf("error: ReadPackedStruct() returned an unexpected error")
	} else if position.X != actualX {
		t.Fatalf("error: expected X = %f, found %f", position.X, actualX)
	} else if position.Y != actualY {
		t.Fatalf("error: expected Y = %f, found %f", position.Y, actualY)
	} else if position.Z != actualZ {
		t.Fatalf("error: expected Z = %f, found %f", position.Z, actualZ)
	}
}
