package weather

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type HistoricalResponse struct {
	Hourly HourlyData `json:"hourly"`
}

type HourlyData struct {
	Time             []string  `json:"time"`
	Temperature2m    []float64 `json:"temperature_2m"`
	WindSpeed10m     []float64 `json:"wind_speed_10m"`
	WindDirection10m []float64 `json:"wind_direction_10m"`
	SolarRadiation   []float64 `json:"shortwave_radiation"`
}

type DataPoint struct {
	Time           time.Time `json:"time"`
	TempAir        float64   `json:"temperature_c"`
	WindSpeed      float64   `json:"wind_speed_mps"`
	WindDir        float64   `json:"wind_direction_deg"`
	SolarRadiation float64   `json:"solar_radiation_wm2"`
}

func FetchRange(lat, lon float64, from, to time.Time) ([]DataPoint, error) {
	return FetchRangeContext(context.Background(), lat, lon, from, to)
}

// FetchRangeContext obtiene datos de archivo y respeta cancelación y deadlines.
func FetchRangeContext(ctx context.Context, lat, lon float64, from, to time.Time) ([]DataPoint, error) {
	return FetchRangeContextWithTimezone(ctx, lat, lon, from, to, "America/Argentina/Buenos_Aires")
}

// FetchRangeContextWithTimezone es la variante configurable de zona horaria.
func FetchRangeContextWithTimezone(ctx context.Context, lat, lon float64, from, to time.Time, timezone string) ([]DataPoint, error) {
	return fetch(ctx, http.DefaultClient, "https://archive-api.open-meteo.com/v1/archive", timezone, lat, lon, from, to)
}

// FetchRangeContextWithCache usa un cache local, opt-in. Un TTL no positivo lo desactiva.
// Los archivos de cache contienen solo la respuesta del proveedor y se escriben con permisos privados.
func FetchRangeContextWithCache(ctx context.Context, lat, lon float64, from, to time.Time, timezone, cacheDir string, ttl time.Duration) ([]DataPoint, error) {
	if strings.TrimSpace(cacheDir) == "" || ttl <= 0 {
		return FetchRangeContextWithTimezone(ctx, lat, lon, from, to, timezone)
	}
	return fetchCached(ctx, http.DefaultClient, "https://archive-api.open-meteo.com/v1/archive", timezone, lat, lon, from, to, cacheDir, ttl)
}

// FetchRangeWithClient es útil para clientes controlados y servidores httptest.
func FetchRangeWithClient(ctx context.Context, client *http.Client, baseURL, timezone string, lat, lon float64, from, to time.Time) ([]DataPoint, error) {
	return fetch(ctx, client, baseURL, timezone, lat, lon, from, to)
}

func fetchCached(ctx context.Context, client *http.Client, baseURL, timezone string, lat, lon float64, from, to time.Time, dir string, ttl time.Duration) ([]DataPoint, error) {
	u, err := requestURL(baseURL, timezone, lat, lon, from, to)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(u.String())))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("crear cache meteorológico: %w", err)
	}
	path := filepath.Join(dir, key+".json")
	if info, err := os.Stat(path); err == nil && time.Since(info.ModTime()) <= ttl {
		body, readErr := os.ReadFile(path)
		if readErr == nil {
			return decodeResponse(body, timezone)
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("crear petición HTTP: %w", err)
	}
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("petición HTTP fallida: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("leer respuesta meteorológica: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".cwind-cache-*")
	if err != nil {
		return nil, fmt.Errorf("crear cache temporal: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(body)
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmpName, path)
	}
	if err != nil {
		return nil, fmt.Errorf("guardar cache meteorológico: %w", err)
	}
	return decodeResponse(body, timezone)
}

func decodeResponse(body []byte, timezone string) ([]DataPoint, error) {
	var apiResp HistoricalResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("decode JSON fallido: %w", err)
	}
	return parseResponse(apiResp, timezone)
}

func parseResponse(apiResp HistoricalResponse, timezone string) ([]DataPoint, error) {
	n := len(apiResp.Hourly.Time)
	if n == 0 {
		return nil, fmt.Errorf("no hay datos horarios en la respuesta")
	}
	if len(apiResp.Hourly.Temperature2m) != n || len(apiResp.Hourly.WindSpeed10m) != n ||
		len(apiResp.Hourly.WindDirection10m) != n || len(apiResp.Hourly.SolarRadiation) != n {
		return nil, fmt.Errorf("respuesta meteorológica incompleta")
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("cargar zona horaria: %w", err)
	}
	result := make([]DataPoint, n)
	for i := range result {
		t, err := time.ParseInLocation("2006-01-02T15:04", apiResp.Hourly.Time[i], loc)
		if err != nil {
			return nil, fmt.Errorf("parsear fecha meteorológica: %w", err)
		}
		result[i] = DataPoint{Time: t, TempAir: apiResp.Hourly.Temperature2m[i],
			WindSpeed: apiResp.Hourly.WindSpeed10m[i] / 3.6, WindDir: apiResp.Hourly.WindDirection10m[i],
			SolarRadiation: apiResp.Hourly.SolarRadiation[i]}
	}
	return result, nil
}

func requestURL(baseURL, timezone string, lat, lon float64, from, to time.Time) (*url.URL, error) {
	if _, err := time.LoadLocation(timezone); err != nil {
		return nil, fmt.Errorf("zona horaria inválida %q: %w", timezone, err)
	}
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, fmt.Errorf("coordenadas fuera de rango")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("fecha final anterior a fecha inicial")
	}
	u, err := url.Parse(baseURL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return nil, fmt.Errorf("URL meteorológica inválida")
	}
	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%.4f", lat))
	q.Set("longitude", fmt.Sprintf("%.4f", lon))
	q.Set("start_date", from.Format("2006-01-02"))
	q.Set("end_date", to.Format("2006-01-02"))
	q.Set("hourly", "temperature_2m,wind_speed_10m,wind_direction_10m,shortwave_radiation")
	q.Set("timezone", timezone)
	u.RawQuery = q.Encode()
	return u, nil
}

func fetch(ctx context.Context, client *http.Client, baseURL, timezone string, lat, lon float64, from, to time.Time) ([]DataPoint, error) {
	if _, err := time.LoadLocation(timezone); err != nil {
		return nil, fmt.Errorf("zona horaria inválida %q: %w", timezone, err)
	}
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, fmt.Errorf("coordenadas fuera de rango")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("fecha final anterior a fecha inicial")
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("URL meteorológica inválida")
	}
	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%.4f", lat))
	q.Set("longitude", fmt.Sprintf("%.4f", lon))
	q.Set("start_date", from.Format("2006-01-02"))
	q.Set("end_date", to.Format("2006-01-02"))
	q.Set("hourly", "temperature_2m,wind_speed_10m,wind_direction_10m,shortwave_radiation")
	q.Set("timezone", timezone)
	u.RawQuery = q.Encode()
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("crear petición HTTP: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("petición HTTP fallida: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var apiResp HistoricalResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode JSON fallido: %w", err)
	}

	n := len(apiResp.Hourly.Time)
	if n == 0 {
		return nil, fmt.Errorf("no hay datos horarios en la respuesta")
	}
	if len(apiResp.Hourly.Temperature2m) != n ||
		len(apiResp.Hourly.WindSpeed10m) != n ||
		len(apiResp.Hourly.WindDirection10m) != n ||
		len(apiResp.Hourly.SolarRadiation) != n {
		return nil, fmt.Errorf("respuesta meteorológica incompleta")
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("cargar zona horaria: %w", err)
	}
	result := make([]DataPoint, n)
	for i := 0; i < n; i++ {
		t, err := time.ParseInLocation("2006-01-02T15:04", apiResp.Hourly.Time[i], loc)
		if err != nil {
			return nil, fmt.Errorf("parsear fecha meteorológica: %w", err)
		}
		result[i] = DataPoint{
			Time:           t,
			TempAir:        apiResp.Hourly.Temperature2m[i],
			WindSpeed:      apiResp.Hourly.WindSpeed10m[i] / 3.6,
			WindDir:        apiResp.Hourly.WindDirection10m[i],
			SolarRadiation: apiResp.Hourly.SolarRadiation[i],
		}
	}
	return result, nil
}
