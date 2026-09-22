package physics

import (
	"math"
	"testing"
)

var stdIEEE738 = Standards["IEEE_738"]

func TestCalculateReturnsPositiveAmpacity(t *testing.T) {
	got := Calculate(WeatherData{
		TempAir: 25, WindSpeed: 2, WindDir: 60, LineAzimuth: 60,
	}, Conductor{
		DiameterMeters: 0.0218, R20: 0.000119, Alpha: 0.004, MaxTemp: 80,
	}, 0, stdIEEE738)
	if got <= 0 || math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("Calculate() = %v, want finite positive value", got)
	}
}

func TestCalculateSolarRadiationReducesAmpacity(t *testing.T) {
	conductor := Conductor{
		DiameterMeters: 0.0218, R20: 0.000119, Alpha: 0.004, MaxTemp: 80,
	}
	base := WeatherData{TempAir: 30, WindSpeed: 1, WindDir: 60, LineAzimuth: 60}
	withoutSolar := Calculate(base, conductor, 0, stdIEEE738)
	withSolar := Calculate(WeatherData{
		TempAir: 30, WindSpeed: 1, WindDir: 60, LineAzimuth: 60,
		SolarRadiation: 900,
	}, conductor, 0, stdIEEE738)
	if withSolar >= withoutSolar {
		t.Fatalf("solar radiation increased ampacity: without=%v with=%v", withoutSolar, withSolar)
	}
}

func TestCalculateHandlesHotAirWithoutNaN(t *testing.T) {
	got := Calculate(WeatherData{
		TempAir: 100, WindSpeed: 0, LineAzimuth: 0,
	}, Conductor{
		DiameterMeters: 0.0218, R20: 0.000119, Alpha: 0.004, MaxTemp: 80,
	}, 0, stdIEEE738)
	if math.IsNaN(got) || math.IsInf(got, 0) || got < 0 {
		t.Fatalf("Calculate() = %v, want finite non-negative value", got)
	}
}

func TestIEEE738ReferenceCase1(t *testing.T) {
	cond := Conductor{
		DiameterMeters: 0.02814, R20: 7.32e-5, Alpha: 0.00403, MaxTemp: 75,
		Emissivity: 0.5, Absorptivity: 0.5,
		SteelCoreDia: 0.0089, AlLayers: 2, AlAreaMM2: 402.8, SteelAreaMM2: 35.5,
	}
	w := WeatherData{
		TempAir: 25, WindSpeed: 0.61, WindDir: 90, LineAzimuth: 0,
		SolarRadiation: 1000, SolarElevation: 90,
	}
	got := Calculate(w, cond, 0, stdIEEE738)
	if got < 700 || got > 1100 {
		t.Errorf("IEEE738 Drake case: got %.0f A, want ~900 A", got)
	}
}

func TestIEEE738ZeroWindNaturalConvection(t *testing.T) {
	cond := Conductor{
		DiameterMeters: 0.0218, R20: 1.19e-4, Alpha: 0.00403, MaxTemp: 80,
		Emissivity: 0.9, Absorptivity: 0.9,
	}
	w := WeatherData{
		TempAir: 40, WindSpeed: 0, WindDir: 0, LineAzimuth: 0,
		SolarRadiation: 0, SolarElevation: 0,
	}
	got := Calculate(w, cond, 0, stdIEEE738)
	if got < 300 || got > 600 {
		t.Errorf("zero wind natural convection: got %.0f A, want 300-600 A", got)
	}
}

func TestWindAngleEffect(t *testing.T) {
	cond := Conductor{
		DiameterMeters: 0.0218, R20: 1.19e-4, Alpha: 0.00403, MaxTemp: 80,
		Emissivity: 0.9, Absorptivity: 0.9,
	}
	wCross := WeatherData{TempAir: 25, WindSpeed: 2, WindDir: 90, LineAzimuth: 0, SolarRadiation: 0, SolarElevation: 0}
	wParallel := WeatherData{TempAir: 25, WindSpeed: 2, WindDir: 0, LineAzimuth: 0, SolarRadiation: 0, SolarElevation: 0}
	ampCross := Calculate(wCross, cond, 0, stdIEEE738)
	ampParallel := Calculate(wParallel, cond, 0, stdIEEE738)
	if ampCross <= ampParallel {
		t.Errorf("cross wind should cool more: cross=%.0f parallel=%.0f", ampCross, ampParallel)
	}
}

func TestACSRBicapaHigherThanHomogeneous(t *testing.T) {
	condACSR := Conductor{
		DiameterMeters: 0.0218, R20: 1.19e-4, Alpha: 0.00403, MaxTemp: 80,
		Emissivity: 0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0072, AlAreaMM2: 240, SteelAreaMM2: 40,
	}
	condHomo := Conductor{
		DiameterMeters: 0.0218, R20: 1.19e-4, Alpha: 0.00403, MaxTemp: 80,
		Emissivity: 0.9, Absorptivity: 0.9,
	}
	w := WeatherData{TempAir: 35, WindSpeed: 0.6, WindDir: 90, LineAzimuth: 0, SolarRadiation: 0, SolarElevation: 0}
	ampACSR := Calculate(w, condACSR, 0, stdIEEE738)
	ampHomo := Calculate(w, condHomo, 0, stdIEEE738)
	if ampACSR <= ampHomo {
		t.Errorf("ACSR bicapa should have higher ampacity: ACSR=%.0f Homo=%.0f", ampACSR, ampHomo)
	}
}

func TestEstimateConductorTempAtMaxTemp(t *testing.T) {
	cond := Conductor{
		DiameterMeters: 0.0218, R20: 1.19e-4, Alpha: 0.00403, MaxTemp: 80,
		Emissivity: 0.9, Absorptivity: 0.9,
	}
	w := WeatherData{TempAir: 25, WindSpeed: 2, WindDir: 90, LineAzimuth: 0, SolarRadiation: 0, SolarElevation: 0}
	amp := Calculate(w, cond, 0, stdIEEE738)
	tc := EstimateConductorTemp(w, cond, amp, 0, stdIEEE738)
	if math.Abs(tc-80.0) > 1.0 {
		t.Errorf("EstimateConductorTemp at DLR ampacity: got %.1f°C, want ~80°C", tc)
	}
}

func TestEstimateConductorTempBelowMax(t *testing.T) {
	cond := Conductor{
		DiameterMeters: 0.0218, R20: 1.19e-4, Alpha: 0.00403, MaxTemp: 80,
		Emissivity: 0.9, Absorptivity: 0.9,
	}
	w := WeatherData{TempAir: 25, WindSpeed: 2, WindDir: 90, LineAzimuth: 0, SolarRadiation: 0, SolarElevation: 0}
	// A la mitad de la ampacidad DLR, la temperatura debe estar bien por debajo de MaxTemp
	amp := Calculate(w, cond, 0, stdIEEE738) * 0.5
	tc := EstimateConductorTemp(w, cond, amp, 0, stdIEEE738)
	if tc >= 80.0 {
		t.Errorf("EstimateConductorTemp at half ampacity: got %.1f°C, want < 80°C", tc)
	}
	if tc <= 25.0 {
		t.Errorf("EstimateConductorTemp at half ampacity: got %.1f°C, want > ambient", tc)
	}
}

func TestEstimateConductorTempZeroCurrent(t *testing.T) {
	cond := Conductor{
		DiameterMeters: 0.0218, R20: 1.19e-4, Alpha: 0.00403, MaxTemp: 80,
		Emissivity: 0.9, Absorptivity: 0.9,
	}
	w := WeatherData{TempAir: 30, WindSpeed: 1, WindDir: 90, LineAzimuth: 0, SolarRadiation: 0, SolarElevation: 0}
	tc := EstimateConductorTemp(w, cond, 0, 0, stdIEEE738)
	if math.Abs(tc-30.0) > 0.1 {
		t.Errorf("EstimateConductorTemp at zero current: got %.1f°C, want ~30°C (ambient)", tc)
	}
}
