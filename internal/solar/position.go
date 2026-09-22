package solar

import (
	"math"
	"time"
)

const (
	deg2rad = math.Pi / 180.0
	rad2deg = 180.0 / math.Pi
)

type Position struct {
	Elevation float64
	Azimuth   float64
	Zenith    float64
}

func Calculate(lat, lon float64, t time.Time) Position {
	utc := t.UTC()

	dayOfYear := utc.YearDay()

	hour := float64(utc.Hour()) + float64(utc.Minute())/60.0 + float64(utc.Second())/3600.0

	gamma := 2.0 * math.Pi * float64(dayOfYear-1) / 365.0
	decl := 0.006918 - 0.399912*math.Cos(gamma) + 0.070257*math.Sin(gamma) - 0.006758*math.Cos(2*gamma) + 0.000907*math.Sin(2*gamma) - 0.002697*math.Cos(3*gamma) + 0.00148*math.Sin(3*gamma)

	eqtime := 229.18 * (0.000075 + 0.001868*math.Cos(gamma) - 0.032077*math.Sin(gamma) - 0.014615*math.Cos(2*gamma) - 0.040849*math.Sin(2*gamma))

	timeOffset := eqtime + 4.0*lon

	tst := hour + timeOffset/60.0

	ha := (tst - 12.0) * 15.0 * deg2rad

	latRad := lat * deg2rad

	sinElev := math.Sin(latRad)*math.Sin(decl) + math.Cos(latRad)*math.Cos(decl)*math.Cos(ha)
	elevRad := math.Asin(sinElev)
	elevation := elevRad * rad2deg

	// Azimut usando fórmula NOAA: atan2(sin(ha), cos(ha)*sin(lat) - tan(decl)*cos(lat))
	// Retorna azimut desde Sur horario: 0=S, 90=O, 180=N, 270=E
	azRad := math.Atan2(math.Sin(ha), math.Cos(ha)*math.Sin(latRad)-math.Tan(decl)*math.Cos(latRad))
	azimuth := azRad * rad2deg
	if azimuth < 0 {
		azimuth += 360.0
	}
	// Convierte a azimut desde Norte horario: 0=N, 90=E, 180=S, 270=O
	azimuth = math.Mod(azimuth+180.0, 360.0)

	zenith := 90.0 - elevation

	return Position{
		Elevation: elevation,
		Azimuth:   azimuth,
		Zenith:    zenith,
	}
}
