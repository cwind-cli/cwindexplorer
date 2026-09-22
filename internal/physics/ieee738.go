package physics

import (
	"math"
)

const (
	stefanBoltzmann = 5.670374419e-8
	airGasConstant  = 287.058
	gravity         = 9.80665
	pi              = math.Pi
)

type WeatherData struct {
	TempAir        float64
	WindSpeed      float64
	WindDir        float64
	LineAzimuth    float64
	SolarRadiation float64
	SolarElevation float64
	PressurePa     float64
}

type Conductor struct {
	DiameterMeters float64
	R20            float64
	Alpha          float64
	MaxTemp        float64
	Emissivity     float64
	Absorptivity   float64
	SteelCoreDia   float64
	AlLayers       int
	AlAreaMM2      float64
	SteelAreaMM2   float64
}

func Calculate(w WeatherData, c Conductor, maxAbsSlopeDeg float64, std StandardConfig) float64 {
	if w.WindSpeed < 0 {
		w.WindSpeed = 0
	}

	diameter := c.DiameterMeters
	emissivity := c.Emissivity
	if emissivity <= 0 {
		emissivity = std.DefaultEmissivity
	}
	absorptivity := c.Absorptivity
	if absorptivity <= 0 {
		absorptivity = std.DefaultAbsorptivity
	}

	windAngle := math.Abs(w.WindDir - w.LineAzimuth)
	if windAngle > 180 {
		windAngle = 360 - windAngle
	}
	windAngleRad := windAngle * pi / 180.0

	filmTemp := (c.MaxTemp + w.TempAir) / 2.0

	rhoAir := airDensity(filmTemp, w.PressurePa)
	muAir := airViscosity(filmTemp)
	kAir := airThermalConductivity(filmTemp)
	cpAir := airSpecificHeat(filmTemp)
	pr := prandtl(rhoAir, muAir, cpAir, kAir)

	reynolds := rhoAir * w.WindSpeed * diameter / muAir

	var nusseltForced float64
	if reynolds <= 0 {
		nusseltForced = 0
	} else {
		nusseltForced = convectionByStandard(std.ConvectionMethod, windAngleRad, reynolds, pr)
	}

	grashof := grashofNumber(diameter, c.MaxTemp-w.TempAir, rhoAir, muAir, filmTemp+273.15, maxAbsSlopeDeg)
	nusseltNatural := nusseltNaturalFlowIEEE738(grashof, pr)

	nusseltCombined := math.Pow(math.Pow(nusseltForced, 4)+math.Pow(nusseltNatural, 4), 0.25)

	hc := nusseltCombined * kAir / diameter
	qc := hc * pi * diameter * (c.MaxTemp - w.TempAir)

	hr := emissivity * stefanBoltzmann * pi * diameter *
		(math.Pow(c.MaxTemp+273.15, 4) - math.Pow(w.TempAir+273.15, 4))

	solar := w.SolarRadiation
	if solar < 0 {
		solar = 0
	}
	solarElevation := w.SolarElevation
	if solarElevation <= 0 {
		solarElevation = 1e-6
	}
	if solarElevation > 90 {
		solarElevation = 90
	}
	qs := absorptivity * solar * diameter * math.Sin(solarElevation*pi/180)

	rAC := acResistanceByStandard(std.ACResistanceMethod, c.R20, c.Alpha, c.MaxTemp, w.TempAir, diameter, c.SteelCoreDia, c.AlAreaMM2, c.SteelAreaMM2, w.PressurePa)

	heatBalance := (qc + hr - qs) / rAC
	if heatBalance <= 0 {
		return 0
	}
	ampacity := math.Sqrt(heatBalance)
	return ampacity
}

func airDensity(tempC float64, pressurePa float64) float64 {
	T := tempC + 273.15
	if pressurePa <= 0 {
		pressurePa = 101325.0
	}
	return pressurePa / (airGasConstant * T)
}

func airViscosity(tempC float64) float64 {
	T := tempC + 273.15
	return 1.458e-6 * math.Pow(T, 1.5) / (T + 110.4)
}

func airThermalConductivity(tempC float64) float64 {
	T := tempC + 273.15
	return 2.42e-3 * math.Pow(T, 1.5) / (T + 194.0)
}

func airSpecificHeat(tempC float64) float64 {
	return 1005.0
}

func prandtl(rho, mu, cp, k float64) float64 {
	return mu * cp / k
}

func kAngleFactor(angleRad float64) float64 {
	angleDeg := angleRad * 180.0 / pi
	if angleDeg > 90 {
		angleDeg = 180 - angleDeg
	}
	switch {
	case angleDeg <= 24:
		return 0.42
	case angleDeg <= 54:
		return 0.58
	case angleDeg <= 84:
		return 0.90
	default:
		return 1.00
	}
}

func nusseltForcedFlowIEEE738(re, pr float64) float64 {
	if re < 0.7 {
		return 0
	}
	if re < 100 {
		return 0.583 * math.Pow(re, 0.471)
	}
	if re < 2650 {
		return 0.641 * math.Pow(re, 0.471)
	}
	return 0.048 * math.Pow(re, 0.8)
}

func grashofNumber(diameter, deltaT, rho, mu, filmTempK, maxAbsSlopeDeg float64) float64 {
	if deltaT < 0 {
		deltaT = 0
	}
	beta := 1.0 / filmTempK
	gr := gravity * beta * deltaT * math.Pow(diameter, 3) * math.Pow(rho, 2) / math.Pow(mu, 2)
	if maxAbsSlopeDeg > 0 {
		slopeDeg := math.Atan(maxAbsSlopeDeg/100.0) * 180.0 / math.Pi
		gr *= math.Cos(slopeDeg * math.Pi / 180.0)
	}
	return gr
}

func nusseltNaturalFlowIEEE738(gr, pr float64) float64 {
	ra := gr * pr
	if ra < 1e4 {
		return 0.42 * math.Pow(ra, 0.25)
	}
	if ra < 1e7 {
		return 0.53 * math.Pow(ra, 0.25)
	}
	if ra < 1e10 {
		return 0.13 * math.Pow(ra, 1.0/3.0)
	}
	return 0.10 * math.Pow(ra, 1.0/3.0)
}

func acResistanceIEEE738(r20, alpha, maxTemp, tempAir, diameter, steelCoreDia float64, alAreaMM2, steelAreaMM2 float64, pressurePa float64) float64 {
	rDC := r20 * (1 + alpha*(maxTemp-20.0))
	if steelCoreDia > 0 && alAreaMM2 > 0 && steelAreaMM2 > 0 {
		skinFactor := skinEffectFactorAl(maxTemp, diameter, steelCoreDia, alAreaMM2)
		proximityFactor := proximityEffectFactor(maxTemp, diameter, alAreaMM2, steelAreaMM2)
		return rDC * skinFactor * proximityFactor
	}
	return rDC * 1.02
}

func skinEffectFactorAl(maxTemp, diameter, steelCoreDia float64, alAreaMM2 float64) float64 {
	if alAreaMM2 <= 0 {
		return 1.0
	}
	f := 50.0
	rhoAl := 28.264 * (1 + 0.00403*(maxTemp-20.0))
	delta := math.Sqrt(rhoAl * 1e-9 / (pi * f * 4e-7))
	if diameter/2 < delta {
		return 1.0
	}
	return 1.0 + 0.01*math.Pow(diameter/2/delta, 2)
}

func proximityEffectFactor(maxTemp, diameter float64, alAreaMM2, steelAreaMM2 float64) float64 {
	if alAreaMM2 <= 0 || steelAreaMM2 <= 0 {
		return 1.0
	}
	steelRatio := steelAreaMM2 / (alAreaMM2 + steelAreaMM2)
	return 1.0 + 0.005*steelRatio*math.Pow(maxTemp/75.0, 2)
}

func EstimateConductorTemp(w WeatherData, c Conductor, current, maxAbsSlopeDeg float64, std StandardConfig) float64 {
	if current <= 0 {
		return w.TempAir
	}
	if w.WindSpeed < 0 {
		w.WindSpeed = 0
	}

	diameter := c.DiameterMeters
	emissivity := c.Emissivity
	if emissivity <= 0 {
		emissivity = std.DefaultEmissivity
	}
	absorptivity := c.Absorptivity
	if absorptivity <= 0 {
		absorptivity = std.DefaultAbsorptivity
	}

	windAngle := math.Abs(w.WindDir - w.LineAzimuth)
	if windAngle > 180 {
		windAngle = 360 - windAngle
	}
	windAngleRad := windAngle * pi / 180.0

	solar := w.SolarRadiation
	if solar < 0 {
		solar = 0
	}
	solarElevation := w.SolarElevation
	if solarElevation <= 0 {
		solarElevation = 1e-6
	}
	if solarElevation > 90 {
		solarElevation = 90
	}
	qs := absorptivity * solar * diameter * math.Sin(solarElevation*pi/180)

	low := w.TempAir
	high := c.MaxTemp + 50.0

	for i := 0; i < 30; i++ {
		mid := (low + high) / 2.0

		filmTempIter := (mid + w.TempAir) / 2.0

		rhoAirIter := airDensity(filmTempIter, w.PressurePa)
		muAirIter := airViscosity(filmTempIter)
		kAirIter := airThermalConductivity(filmTempIter)
		cpAirIter := airSpecificHeat(filmTempIter)
		prIter := prandtl(rhoAirIter, muAirIter, cpAirIter, kAirIter)

		reynoldsIter := rhoAirIter * w.WindSpeed * diameter / muAirIter

		var nusseltForcedIter float64
		if reynoldsIter <= 0 {
			nusseltForcedIter = 0
		} else {
			nusseltForcedIter = convectionByStandard(std.ConvectionMethod, windAngleRad, reynoldsIter, prIter)
		}

		grashof := grashofNumber(diameter, mid-w.TempAir, rhoAirIter, muAirIter, filmTempIter+273.15, maxAbsSlopeDeg)
		nusseltNatural := nusseltNaturalFlowIEEE738(grashof, prIter)

		nusseltCombined := math.Pow(math.Pow(nusseltForcedIter, 4)+math.Pow(nusseltNatural, 4), 0.25)

		hc := nusseltCombined * kAirIter / diameter
		qc := hc * pi * diameter * (mid - w.TempAir)

		hr := emissivity * stefanBoltzmann * pi * diameter *
			(math.Pow(mid+273.15, 4) - math.Pow(w.TempAir+273.15, 4))

		rAC := acResistanceByStandard(std.ACResistanceMethod, c.R20, c.Alpha, mid, w.TempAir, diameter, c.SteelCoreDia, c.AlAreaMM2, c.SteelAreaMM2, w.PressurePa)

		heatGen := current * current * rAC
		heatDiss := qc + hr - qs

		if heatGen > heatDiss {
			low = mid
		} else {
			high = mid
		}
	}

	return (low + high) / 2.0
}
