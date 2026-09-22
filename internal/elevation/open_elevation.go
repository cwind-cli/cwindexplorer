package elevation

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cwind-cli/cwind/internal/geo"
)

const (
	openElevationURL = "https://api.open-elevation.com/api/v1/lookup"
	timeout          = 10 * time.Second
)

type openElevationResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Elevation float64 `json:"elevation"`
	} `json:"results"`
}

func GetElevation(lat, lon float64) (float64, error) {
	u, _ := url.Parse(openElevationURL)
	q := u.Query()
	q.Set("locations", fmt.Sprintf("%.6f,%.6f", lat, lon))
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(u.String())
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data openElevationResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}
	if len(data.Results) == 0 {
		return 0, fmt.Errorf("no elevation data")
	}
	return data.Results[0].Elevation, nil
}

func GetElevationBatch(coords []geo.Coordinate) ([]float64, error) {
	if len(coords) == 0 {
		return nil, nil
	}

	var locations []string
	for _, c := range coords {
		locations = append(locations, fmt.Sprintf("%.6f,%.6f", c.Lat, c.Lon))
	}

	u, _ := url.Parse(openElevationURL)
	q := u.Query()
	q.Set("locations", strings.Join(locations, "|"))
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data openElevationResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	elevs := make([]float64, len(coords))
	for i, r := range data.Results {
		elevs[i] = r.Elevation
	}
	return elevs, nil
}

func PressureFromElevation(elevM float64) float64 {
	const (
		P0       = 101325.0
		L        = 0.0065
		T0       = 288.15
		exponent = 5.25588
	)
	if elevM < 0 {
		elevM = 0
	}
	return P0 * math.Pow(1.0-L*elevM/T0, exponent)
}
