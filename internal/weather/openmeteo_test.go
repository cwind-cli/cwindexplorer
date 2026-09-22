package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestFetchRangeWithClient(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("timezone") != "UTC" {
			t.Errorf("timezone missing")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hourly":{"time":["2024-01-01T00:00"],"temperature_2m":[20],"wind_speed_10m":[36],"wind_direction_10m":[90],"shortwave_radiation":[0]}}`))
	}))
	defer s.Close()
	got, err := FetchRangeWithClient(context.Background(), s.Client(), s.URL, "UTC", 0, 0, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil || len(got) != 1 || got[0].WindSpeed != 10 {
		t.Fatalf("unexpected result: %v %#v", err, got)
	}
}

func TestFetchCachedUsesLocalResponse(t *testing.T) {
	requests := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"hourly":{"time":["2024-01-01T00:00"],"temperature_2m":[20],"wind_speed_10m":[36],"wind_direction_10m":[90],"shortwave_radiation":[0]}}`))
	}))
	defer s.Close()
	dir := t.TempDir()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := fetchCached(context.Background(), s.Client(), s.URL, "UTC", 0, 0, from, from, dir, time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := fetchCached(context.Background(), s.Client(), s.URL, "UTC", 0, 0, from, from, dir, time.Hour); err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("expected one network request, got %d", requests)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected one cache file, got %d", len(entries))
	}
}
