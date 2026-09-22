package solar

import (
	"testing"
	"time"
)

func TestSolarPosition(t *testing.T) {
	tests := []struct {
		name      string
		lat, lon  float64
		t         time.Time
		wantElev  float64
		wantAzim  float64
		tolerance float64
	}{
		{
			name:      "Buenos Aires summer solar noon",
			lat:       -34.6,
			lon:       -58.4,
			t:         time.Date(2024, 1, 1, 15, 56, 0, 0, time.UTC), // solar noon
			wantElev:  78.0,
			wantAzim:  0.0,
			tolerance: 2.0,
		},
		{
			name:      "Buenos Aires winter solar noon",
			lat:       -34.6,
			lon:       -58.4,
			t:         time.Date(2024, 7, 1, 15, 30, 0, 0, time.UTC), // solar noon (approx)
			wantElev:  31.0,
			wantAzim:  0.0,
			tolerance: 8.0,
		},
		{
			name:      "Equator equinox solar noon",
			lat:       0.0,
			lon:       0.0,
			t:         time.Date(2024, 3, 20, 12, 0, 0, 0, time.UTC),
			wantElev:  88.0,
			wantAzim:  90.0,
			tolerance: 5.0,
		},
		{
			name:      "Buenos Aires summer morning",
			lat:       -34.6,
			lon:       -58.4,
			t:         time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), // 11:04 local
			wantElev:  65.0,
			wantAzim:  50.0,
			tolerance: 10.0,
		},
		{
			name:      "Night time (negative elevation)",
			lat:       -34.6,
			lon:       -58.4,
			t:         time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC), // 03:00 local
			wantElev:  -15.0,
			wantAzim:  150.0,
			tolerance: 15.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pos := Calculate(tc.lat, tc.lon, tc.t)
			if diff := pos.Elevation - tc.wantElev; diff > tc.tolerance || diff < -tc.tolerance {
				t.Errorf("elevation: got %.2f, want %.2f ±%.1f", pos.Elevation, tc.wantElev, tc.tolerance)
			}
			if diff := pos.Azimuth - tc.wantAzim; diff > tc.tolerance || diff < -tc.tolerance {
				t.Errorf("azimuth: got %.2f, want %.2f ±%.1f", pos.Azimuth, tc.wantAzim, tc.tolerance)
			}
		})
	}
}

func TestSolarPositionZenith(t *testing.T) {
	pos := Calculate(-34.6, -58.4, time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC))
	if pos.Zenith != 90-pos.Elevation {
		t.Errorf("zenith should be 90-elevation: got zenith=%.2f, elev=%.2f", pos.Zenith, pos.Elevation)
	}
}
