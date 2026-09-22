package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cwind-cli/cwind/internal/weather"
)

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), err
}

func TestReportCLIJSONAndCSVOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.csv")
	if err := os.WriteFile(path, []byte("lat,lon\n-34.60,-58.40\n-34.61,-58.39\n"), 0600); err != nil {
		t.Fatal(err)
	}
	old := fetchWeather
	fetchWeather = func(_ context.Context, _, _ float64, _, _ time.Time, _, _ string, _ time.Duration) ([]weather.DataPoint, error) {
		return []weather.DataPoint{{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), TempAir: 25, WindSpeed: 2, WindDir: 90}}, nil
	}
	defer func() { fetchWeather = old }()
	out, err := runCLI(t, "report", "--trace", path, "--conductor", "acsr-795-drake", "--from", "2024-01-01", "--to", "2024-01-01", "--format", "json")
	if err != nil || !strings.Contains(out, `"status": "ok"`) {
		t.Fatalf("unexpected JSON report: %v %q", err, out)
	}
	out, err = runCLI(t, "report", "--trace", path, "--conductor", "acsr-795-drake", "--from", "2024-01-01", "--to", "2024-01-01", "--format", "csv")
	if err != nil || !strings.Contains(out, "# cwind_schema_version") || !strings.Contains(out, "ampacity_a") {
		t.Fatalf("unexpected CSV report: %v %q", err, out)
	}
}

func TestReportCLIError(t *testing.T) {
	_, err := runCLI(t, "report", "--trace", "missing.csv", "--conductor", "acsr-795-drake", "--from", "2024-01-01", "--to", "2024-01-01", "--format", "json")
	if err == nil || !strings.Contains(err.Error(), "traza:") {
		t.Fatalf("expected report error, got %v", err)
	}
}

func TestValidateCLIOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.csv")
	if err := os.WriteFile(path, []byte("lat,lon\n-34.60,-58.40\n-34.61,-58.39\n"), 0600); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "validate", "--trace", path, "--conductor", "acsr-795-drake")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if !strings.Contains(out, "válido: 2 puntos") {
		t.Fatalf("unexpected validate output: %q", out)
	}
}

func TestValidateCLIError(t *testing.T) {
	_, err := runCLI(t, "validate", "--trace", filepath.Join(t.TempDir(), "missing.csv"))
	if err == nil || !strings.Contains(err.Error(), "abrir archivo") {
		t.Fatalf("expected useful validation error, got %v", err)
	}
}
