package geo

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Coordinate struct {
	Lat float64
	Lon float64
}

func ParseFile(path string) (Coordinate, error) {
	trace, err := ParseTraceFile(path)
	if err != nil {
		return Coordinate{}, err
	}
	return centroid(trace), nil
}

func parseTraceFile(path string) (Trace, error) {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(path)))
	if ext != ".csv" && ext != ".kml" {
		return nil, fmt.Errorf("formato no compatible: use un archivo .csv o .kml")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("abrir archivo: %w", err)
	}
	defer f.Close()

	if ext == ".kml" {
		return parseKMLTrace(f)
	}
	return parseCSVTrace(f)
}

func parseKML(r io.Reader) (Coordinate, error) {
	coords, err := parseKMLTrace(r)
	if err != nil {
		return Coordinate{}, err
	}
	return centroid(coords), nil
}

func parseKMLTrace(r io.Reader) (Trace, error) {
	dec := xml.NewDecoder(r)
	var coords []Coordinate

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsear KML: %w", err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "coordinates" {
			var text string
			if err := dec.DecodeElement(&text, &se); err != nil {
				return nil, fmt.Errorf("leer coordenadas KML: %w", err)
			}
			coords = parseCoordString(text)
			break
		}
	}

	if len(coords) == 0 {
		return nil, fmt.Errorf("no se encontraron coordenadas en KML")
	}
	return coords, nil
}

func parseCSV(r io.Reader) (Coordinate, error) {
	coords, err := parseCSVTrace(r)
	if err != nil {
		return Coordinate{}, err
	}
	return centroid(coords), nil
}

func parseCSVTrace(r io.Reader) (Trace, error) {
	reader := csv.NewReader(r)
	reader.Comma = ','
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("leer CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV vacío o solo cabecera")
	}

	latIndex, lonIndex := -1, -1
	for i, column := range records[0] {
		switch strings.ToLower(strings.TrimSpace(column)) {
		case "lat", "latitude", "latitud", "latitud_dd":
			latIndex = i
		case "lon", "lng", "longitude", "longitud", "longitud_dd":
			lonIndex = i
		}
	}
	if latIndex < 0 || lonIndex < 0 {
		return nil, fmt.Errorf("CSV: la cabecera debe incluir columnas lat,lon o Latitud_DD,Longitud_DD")
	}

	var coords []Coordinate
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) <= latIndex || len(row) <= lonIndex {
			return nil, fmt.Errorf("CSV: fila %d debe contener latitud y longitud", i+1)
		}
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(row[latIndex]), 64)
		lon, err2 := strconv.ParseFloat(strings.TrimSpace(row[lonIndex]), 64)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("CSV: fila %d debe contener latitud y longitud numéricas", i+1)
		}
		if lat < -90 || lat > 90 {
			return nil, fmt.Errorf("CSV: latitud inválida en fila %d (debe estar entre -90 y 90)", i+1)
		}
		if lon < -180 || lon > 180 {
			return nil, fmt.Errorf("CSV: longitud inválida en fila %d (debe estar entre -180 y 180)", i+1)
		}
		coords = append(coords, Coordinate{Lat: lat, Lon: lon})
	}

	if len(coords) == 0 {
		return nil, fmt.Errorf("no hay coordenadas válidas en CSV")
	}
	return coords, nil
}

func parseCoordString(s string) []Coordinate {
	var result []Coordinate
	parts := strings.Fields(s)
	for _, p := range parts {
		xy := strings.Split(p, ",")
		if len(xy) < 2 {
			continue
		}
		lon, err1 := strconv.ParseFloat(xy[0], 64)
		lat, err2 := strconv.ParseFloat(xy[1], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
			continue
		}
		result = append(result, Coordinate{Lat: lat, Lon: lon})
	}
	return result
}

func centroid(coords []Coordinate) Coordinate {
	var sumLat, sumLon float64
	for _, c := range coords {
		sumLat += c.Lat
		sumLon += c.Lon
	}
	n := float64(len(coords))
	return Coordinate{Lat: sumLat / n, Lon: sumLon / n}
}
