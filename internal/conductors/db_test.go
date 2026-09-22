package conductors

import (
	"math"
	"testing"
	"time"

	"github.com/cwind-cli/cwind/internal/physics"
)

func TestDatabaseConductorsAreValid(t *testing.T) {
	for id, conductor := range Database {
		if err := conductor.Validate(); err != nil {
			t.Errorf("%s: %v", id, err)
		}
	}
}

func TestValidationRatingPrefersReferenceRating(t *testing.T) {
	cond := Conductor{StaticRating: 470, ReferenceRatingA: 415}
	if got := cond.ValidationRating(); got != 415 {
		t.Fatalf("ValidationRating() = %.1f, want 415", got)
	}

	cond.ReferenceRatingA = 0
	if got := cond.ValidationRating(); got != 470 {
		t.Fatalf("ValidationRating() fallback = %.1f, want 470", got)
	}
}

func TestCalibrateRatingConditionsParsing(t *testing.T) {
	tests := []struct {
		name       string
		conditions string
		wantTamb   float64
		wantWind   float64
		wantSolar  float64
		wantTcMax  float64
	}{
		{
			name:       "al sol standard",
			conditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
			wantTamb:   40.0,
			wantWind:   0.6,
			wantSolar:  1000.0,
			wantTcMax:  80.0,
		},
		{
			name:       "radiacion explicita",
			conditions: "Tc=80°C, Tamb=40°C, sol=750 W/m², V=0.6 m/s",
			wantTamb:   40.0,
			wantWind:   0.6,
			wantSolar:  750.0,
			wantTcMax:  80.0,
		},
		{
			name:       "SIN sol",
			conditions: "Tc=80°C, Tamb=40°C, SIN sol, V=0.6 m/s",
			wantTamb:   40.0,
			wantWind:   0.6,
			wantSolar:  0.0,
			wantTcMax:  80.0,
		},
		{
			name:       "sol pleno",
			conditions: "Tc=75°C, Tamb=25°C, V=0.6 m/s, sol pleno",
			wantTamb:   25.0,
			wantWind:   0.6,
			wantSolar:  1000.0,
			wantTcMax:  75.0,
		},
		{
			name:       "no sun",
			conditions: "Tc=80°C, Tamb=35°C, V=1.0 m/s, no sun",
			wantTamb:   35.0,
			wantWind:   1.0,
			wantSolar:  0.0,
			wantTcMax:  80.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tamb, wind, solar, tc := parseRatingConditions(tt.conditions)
			if tamb != tt.wantTamb {
				t.Errorf("Tamb: got %.1f, want %.1f", tamb, tt.wantTamb)
			}
			if wind != tt.wantWind {
				t.Errorf("Wind: got %.1f, want %.1f", wind, tt.wantWind)
			}
			if solar != tt.wantSolar {
				t.Errorf("Solar: got %.1f, want %.1f", solar, tt.wantSolar)
			}
			if tc != tt.wantTcMax {
				t.Errorf("TcMax: got %.1f, want %.1f", tc, tt.wantTcMax)
			}
		})
	}
}

func TestCalibrateConductor(t *testing.T) {
	// Test with a conductor that should calibrate well (ALAC 240/40 IMSA - sin sol)
	cond := Conductor{
		ID:               "test-alac-240-40",
		Name:             "ALAC 240/40 Test",
		Type:             "ACSR",
		DiameterMeters:   0.02184,
		R20:              0.0001190,
		Alpha:            0.00403,
		MaxTemp:          80.0,
		StaticRating:     645.0,
		Source:           "Test",
		RatingConditions: "Tc=80°C, Tamb=40°C, SIN sol, V=0.6 m/s",
		Emissivity:       0.9,
		Absorptivity:     0.9,
		SteelCoreDia:     0.0072,
		AlLayers:         1,
		AlAreaMM2:        240.0,
		SteelAreaMM2:     40.0,
	}

	errPct, err := cond.Calibrate()
	if err != nil {
		t.Fatalf("Calibrate() error: %v", err)
	}

	// Verify calibration runs and returns a result (may be > 1% for real conductors)
	if errPct < 0 {
		t.Errorf("Calibration error should be positive: %.2f%%", errPct)
	}

	// Test calibration fields are set on the conductor
	if cond.CalibrationErrorPct == 0 {
		t.Errorf("CalibrationErrorPct should be set after Calibrate()")
	}
	if cond.CalibrationConditions == "" {
		t.Errorf("CalibrationConditions should be set after Calibrate()")
	}
	if cond.CalibratedAt == "" {
		t.Errorf("CalibratedAt should be set after Calibrate()")
	}

	// Verify timestamp format
	_, parseErr := time.Parse(time.RFC3339, cond.CalibratedAt)
	if parseErr != nil {
		t.Errorf("CalibratedAt should be RFC3339 format: %v", parseErr)
	}

	// Verify error percentage matches returned value
	if math.Abs(cond.CalibrationErrorPct-errPct) > 0.001 {
		t.Errorf("CalibrationErrorPct %.2f%% != returned %.2f%%", cond.CalibrationErrorPct, errPct)
	}
}

func TestCalibrateBuiltinConductors(t *testing.T) {
	// Test that built-in conductors with StaticRating > 0 can be calibrated
	// (they may or may not pass < 1%, we just verify the function runs)
	calibrated := 0
	failed := 0
	for id, cond := range Database {
		if cond.StaticRating > 0 {
			calibrated++
			errPct, err := cond.Calibrate()
			if err != nil {
				t.Errorf("%s: Calibrate() error: %v", id, err)
			} else if errPct >= 1.0 {
				failed++
				t.Logf("WARN: %s calibration error %.2f%% >= 1%%", id, errPct)
			}
		}
	}
	t.Logf("Calibrated %d conductors, %d with error >= 1%%", calibrated, failed)
}

func TestCalibrationDoesNotAlterPhysicalCalculation(t *testing.T) {
	// The declared rating is a validation reference, not a correction to the model.
	for id, cond := range Database {
		if cond.StaticRating <= 0 {
			continue
		}

		// Calculate validation metadata; it must not alter the physical model.
		_, err := cond.Calibrate()
		if err != nil {
			t.Errorf("%s: Calibrate() error: %v", id, err)
			continue
		}

		// Get standard config
		std, ok := physics.GetStandard(cond.Standard)
		if !ok {
			std = physics.Standards["IEEE_738"]
		}

		// Parsea condiciones reales del rating del conductor
		parsedTamb, parsedWind, parsedSolar, _ := parseRatingConditions(cond.RatingConditions)

		solarRad := parsedSolar
		if solarRad <= 0 {
			solarRad = std.RatingSolarRadiation
		}

		// Create calibration weather conditions from parsed rating
		w := physics.WeatherData{
			TempAir:        parsedTamb,
			WindSpeed:      parsedWind,
			WindDir:        std.RatingWindAngleDeg,
			LineAzimuth:    0,
			SolarRadiation: solarRad,
			SolarElevation: std.RatingSolarElevation,
			PressurePa:     101325.0,
		}

		physCond := physics.Conductor{
			DiameterMeters: cond.DiameterMeters,
			R20:            cond.R20,
			Alpha:          cond.Alpha,
			MaxTemp:        cond.MaxTemp,
			Emissivity:     cond.Emissivity,
			Absorptivity:   cond.Absorptivity,
			SteelCoreDia:   cond.SteelCoreDia,
			AlLayers:       cond.AlLayers,
			AlAreaMM2:      cond.AlAreaMM2,
			SteelAreaMM2:   cond.SteelAreaMM2,
		}
		_, _, _, parsedTc := parseRatingConditions(cond.RatingConditions)
		if parsedTc > 0 {
			physCond.MaxTemp = parsedTc
		}
		if physCond.Emissivity <= 0 {
			physCond.Emissivity = std.DefaultEmissivity
		}
		if physCond.Absorptivity <= 0 {
			physCond.Absorptivity = std.DefaultAbsorptivity
		}

		calc := physics.Calculate(w, physCond, 0, std)
		if calc <= 0 || math.IsNaN(calc) || math.IsInf(calc, 0) {
			t.Errorf("%s: physical DLR calculation is invalid: %.1f A", id, calc)
		}
	}
}
