package geo

import "math"

// Trace es el conjunto ordenado de vértices parseados desde un archivo de línea.
type Trace []Coordinate

// Azimuth retorna el rumbo inicial (grados horario desde el norte) de una traza.
func (t Trace) Azimuth() float64 {
	if len(t) < 2 {
		return 0
	}
	a, b := t[0], t[len(t)-1]
	lat1, lat2 := a.Lat*math.Pi/180, b.Lat*math.Pi/180
	dlon := (b.Lon - a.Lon) * math.Pi / 180
	y := math.Sin(dlon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dlon)
	return math.Mod(math.Atan2(y, x)*180/math.Pi+360, 360)
}

// Elevations almacena la elevación (metros sobre nivel del mar) para cada vértice.
type Elevations []float64

// Slopes almacena el porcentaje de pendiente entre vértices consecutivos.
// Slopes[i] es la pendiente del vértice i al i+1.
// Positivo = subida, Negativo = bajada.
type Slopes []float64

// TraceData contiene la traza parseada con datos de elevación y pendiente.
type TraceData struct {
	Coordinates Trace
	Elevations  Elevations
	Slopes      Slopes
	MaxAbsSlope float64 // Porcentaje de pendiente absoluta máxima
	MaxAbsIndex int     // Índice del segmento con pendiente absoluta máxima
}

// ParseTraceFile parsea todos los vértices desde un archivo.
// Para datos de elevación, use ParseTraceFileWithElevation.
func ParseTraceFile(path string) (Trace, error) {
	return parseTraceFile(path)
}

// ParseTraceFileSimple parsea vértices sin datos de elevación (compatibilidad hacia atrás).
func ParseTraceFileSimple(path string) (Trace, error) {
	return parseTraceFile(path)
}

// ParseTraceFileWithElevation parsea todos los vértices y calcula pendientes desde elevaciones provistas.
func ParseTraceFileWithElevation(path string, elevs []float64) (TraceData, error) {
	coords, err := parseTraceFile(path)
	if err != nil {
		return TraceData{}, err
	}

	if len(elevs) != len(coords) {
		elevs = make([]float64, len(coords))
	}

	slopes := calculateSlopes(coords, elevs)
	maxAbsSlope, maxAbsIndex := findMaxAbsSlope(slopes)

	return TraceData{
		Coordinates: coords,
		Elevations:  elevs,
		Slopes:      slopes,
		MaxAbsSlope: maxAbsSlope,
		MaxAbsIndex: maxAbsIndex,
	}, nil
}

func calculateSlopes(coords []Coordinate, elevs []float64) []float64 {
	if len(coords) < 2 || len(elevs) != len(coords) {
		return []float64{}
	}
	slopes := make([]float64, len(coords)-1)
	for i := 0; i < len(coords)-1; i++ {
		dist := HaversineDistance(coords[i], coords[i+1])
		if dist == 0 {
			slopes[i] = 0
			continue
		}
		deltaElev := elevs[i+1] - elevs[i]
		slopes[i] = (deltaElev / dist) * 100.0 // percentage
	}
	return slopes
}

func findMaxAbsSlope(slopes []float64) (float64, int) {
	if len(slopes) == 0 {
		return 0, -1
	}
	maxAbs := math.Abs(slopes[0])
	maxIdx := 0
	for i, s := range slopes {
		abs := math.Abs(s)
		if abs > maxAbs {
			maxAbs = abs
			maxIdx = i
		}
	}
	return maxAbs, maxIdx
}

// HaversineDistance calcula la distancia de gran círculo entre dos coordenadas en metros.
func HaversineDistance(a, b Coordinate) float64 {
	const R = 6371000.0 // Radio de la Tierra en metros
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLon := (b.Lon - a.Lon) * math.Pi / 180

	sinDLat := math.Sin(dLat / 2)
	sinDLon := math.Sin(dLon / 2)
	h := sinDLat*sinDLat + math.Cos(lat1)*math.Cos(lat2)*sinDLon*sinDLon
	return 2 * R * math.Asin(math.Sqrt(h))
}

// Azimuth retorna el rumbo inicial (grados horario desde el norte) de una traza.
func (t TraceData) Azimuth() float64 {
	if len(t.Coordinates) < 2 {
		return 0
	}
	a, b := t.Coordinates[0], t.Coordinates[len(t.Coordinates)-1]
	lat1, lat2 := a.Lat*math.Pi/180, b.Lat*math.Pi/180
	dlon := (b.Lon - a.Lon) * math.Pi / 180
	y := math.Sin(dlon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dlon)
	return math.Mod(math.Atan2(y, x)*180/math.Pi+360, 360)
}

// Para compatibilidad hacia atrás - TraceData sigue funcionando como []Coordinate
func (t TraceData) Vertices() []Coordinate {
	return t.Coordinates
}
