package geo

import (
	"math"
	"testing"
)

func TestTraceAzimuthNorth(t *testing.T) {
	a := (Trace{{Lat: 0, Lon: 0}, {Lat: 1, Lon: 0}}).Azimuth()
	if math.Abs(a) > 0.01 {
		t.Fatalf("azimuth = %v, want north", a)
	}
}
