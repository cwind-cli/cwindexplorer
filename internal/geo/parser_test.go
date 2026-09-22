package geo

import (
	"strings"
	"testing"
)

func TestParseCSVWithSimpleHeaders(t *testing.T) {
	got, err := parseCSV(strings.NewReader("lat,lon\n-27.0,-55.0\n-29.0,-57.0\n"))
	if err != nil {
		t.Fatalf("parseCSV() error = %v", err)
	}
	if got.Lat != -28 || got.Lon != -56 {
		t.Fatalf("parseCSV() = %+v, want {-28 -56}", got)
	}
}

func TestParseCSVWithProductionHeaders(t *testing.T) {
	input := "ID,Latitud_DD,Longitud_DD\nV1,-27.0,-55.0\nV2,-28.0,-56.0\n"
	got, err := parseCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseCSV() error = %v", err)
	}
	if got.Lat != -27.5 || got.Lon != -55.5 {
		t.Fatalf("parseCSV() = %+v, want {-27.5 -55.5}", got)
	}
}

func TestParseCSVRejectsInvalidCoordinate(t *testing.T) {
	_, err := parseCSV(strings.NewReader("lat,lon\n-91,-55\n"))
	if err == nil {
		t.Fatal("parseCSV() accepted an invalid latitude")
	}
}

func TestParseKML(t *testing.T) {
	input := `<kml><coordinates>-55,-27,0 -57,-29,0</coordinates></kml>`
	got, err := parseKML(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseKML() error = %v", err)
	}
	if got.Lat != -28 || got.Lon != -56 {
		t.Fatalf("parseKML() = %+v, want {-28 -56}", got)
	}
}
