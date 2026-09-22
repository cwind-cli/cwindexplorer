package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cwind-cli/cwind/internal/conductors"
	"github.com/cwind-cli/cwind/internal/elevation"
	"github.com/cwind-cli/cwind/internal/geo"
	"github.com/cwind-cli/cwind/internal/physics"
	"github.com/cwind-cli/cwind/internal/solar"
	"github.com/cwind-cli/cwind/internal/version"
	"github.com/cwind-cli/cwind/internal/weather"
	"github.com/spf13/cobra"
)

type reportResult struct {
	weather.DataPoint
	Ampacity     float64 `json:"ampacity_a"`
	TcEst        float64 `json:"tc_est_c,omitempty"`
	TcEstAtLimit float64 `json:"tc_est_at_limit_c,omitempty"`
}
type reportDocument struct {
	SchemaVersion          string         `json:"schema_version"`
	Status                 string         `json:"status"`
	Errors                 []string       `json:"errors,omitempty"`
	CwindVersion           string         `json:"cwind_version"`
	WeatherSource          string         `json:"weather_source"`
	Model                  string         `json:"model"`
	Conductor              string         `json:"conductor"`
	Latitude               float64        `json:"latitude"`
	Longitude              float64        `json:"longitude"`
	ConductorID            string         `json:"conductor_id"`
	LineAzimuth            float64        `json:"line_azimuth_deg"`
	Timezone               string         `json:"timezone"`
	LimiteAdminMinA        float64        `json:"limite_admin_min_a,omitempty"`
	LimiteAdminMaxA        float64        `json:"limite_admin_max_a,omitempty"`
	LimiteConductorA       float64        `json:"limite_conductor_a,omitempty"`
	RatingReferenceA       float64        `json:"rating_reference_a,omitempty"`
	ValidationCalculatedA  float64        `json:"validation_calculated_a,omitempty"`
	ValidationDeviationPct float64        `json:"validation_deviation_pct,omitempty"`
	ElevationAvgM          float64        `json:"elevation_avg_m,omitempty"`
	Results                []reportResult `json:"results"`
}

var (
	styleHeader     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d47e")).Bold(true)
	styleSubtle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7c8a9e"))
	styleTitle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d47e")).Bold(true)
	stylePurple     = lipgloss.NewStyle().Foreground(lipgloss.Color("#b794f4"))
	styleCyan       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d4d4"))
	styleRiskNormal = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d47e"))
	styleRiskAlert  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b35")).Bold(true)
	styleLimitOp    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b35")).Bold(true)
	styleLimitCond  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d47e"))
	styleHeaderLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#2a3440"))
	styleMuted      = lipgloss.NewStyle().Foreground(lipgloss.Color("#5c6a7e"))
)

var fetchWeather = weather.FetchRangeContextWithCache

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Genera un reporte horario de ampacidad dinámica (DLR)",
	Args:  cobra.NoArgs,
	RunE:  runReport,
}

func runReport(cmd *cobra.Command, args []string) error {
	tracePath, _ := cmd.Flags().GetString("trace")
	id, _ := cmd.Flags().GetString("conductor")
	lineDeprecated, _ := cmd.Flags().GetString("line")
	if lineDeprecated != "" {
		if id != "" {
			return fmt.Errorf("usá --conductor o --line, no ambos")
		}
		fmt.Fprintln(cmd.ErrOrStderr(), "aviso: --line está deprecado para seleccionar conductores, usá --conductor.\n   --line se usará en el futuro para trazas geográficas.")
		id = lineDeprecated
	}
	fromText, _ := cmd.Flags().GetString("from")
	toText, _ := cmd.Flags().GetString("to")
	azimuth, _ := cmd.Flags().GetFloat64("line-azimuth")
	azimuthSet := cmd.Flags().Changed("line-azimuth")
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	timezone, _ := cmd.Flags().GetString("timezone")
	cacheDir, _ := cmd.Flags().GetString("cache-dir")
	cacheTTL, _ := cmd.Flags().GetDuration("cache-ttl")
	showTech, _ := cmd.Flags().GetBool("detalle-tecnico")
	adminLimitMin, _ := cmd.Flags().GetFloat64("admin-min")
	adminLimitMax, _ := cmd.Flags().GetFloat64("admin-max")
	if strings.TrimSpace(tracePath) == "" {
		return fmt.Errorf("falta --trace")
	}
	if format != "text" && format != "json" && format != "csv" {
		return fmt.Errorf("--format debe ser text, json o csv")
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("--timezone inválida: %w", err)
	}
	from, err := time.ParseInLocation("2006-01-02", fromText, location)
	if err != nil {
		return fmt.Errorf("--from debe tener formato YYYY-MM-DD")
	}
	to, err := time.ParseInLocation("2006-01-02", toText, from.Location())
	if err != nil {
		return fmt.Errorf("--to debe tener formato YYYY-MM-DD")
	}
	if to.Before(from) {
		return fmt.Errorf("--to debe ser posterior o igual a --from")
	}
	hours := to.Sub(from).Hours()
	if hours >= 24 {
		return fmt.Errorf("error: rango máximo es 24h (pediste %.0fh). Reducí el rango con --from/--to.", hours)
	}
	trace, err := geo.ParseTraceFile(tracePath)
	if err != nil {
		return fmt.Errorf("traza: %w", err)
	}

	// Obtiene elevaciones para todos los vértices
	elevs, err := elevation.GetElevationBatch(trace)
	if err != nil {
		elevs = make([]float64, len(trace))
	}

	traceData, err := geo.ParseTraceFileWithElevation(tracePath, elevs)
	if err != nil {
		return fmt.Errorf("traza: %w", err)
	}

	coord := geo.Coordinate{Lat: 0, Lon: 0}
	for _, p := range traceData.Coordinates {
		coord.Lat += p.Lat
		coord.Lon += p.Lon
	}
	coord.Lat /= float64(len(traceData.Coordinates))
	coord.Lon /= float64(len(traceData.Coordinates))
	if !azimuthSet {
		azimuth = traceData.Azimuth()
	}
	if azimuth < 0 || azimuth >= 360 {
		return fmt.Errorf("--line-azimuth debe estar entre 0 y 360 grados")
	}
	conductors.EnsureCalibration()
	cond, ok := conductors.GetMerged(id)
	if id == "" {
		cond, id, err = chooseConductor()
		ok = err == nil
	}
	if !ok {
		if err != nil {
			return err
		}
		return fmt.Errorf("conductor %q no encontrado", id)
	}
	if err := cond.Validate(); err != nil {
		return err
	}
	data, err := fetchWeather(cmd.Context(), coord.Lat, coord.Lon, from, to, timezone, cacheDir, cacheTTL)
	if err != nil {
		return fmt.Errorf("weather: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("la respuesta meteorológica no contiene datos")
	}

	elev, err := elevation.GetElevation(coord.Lat, coord.Lon)
	if err != nil {
		elev = 0
	}
	pressure := elevation.PressureFromElevation(elev)

	physCond := physics.Conductor{
		DiameterMeters: cond.DiameterMeters,
		R20:            cond.R20,
		Alpha:          cond.Alpha,
		MaxTemp:        cond.MaxTemp,
		Emissivity:     cond.Emissivity,
		Absorptivity:   cond.Absorptivity,
		SteelCoreDia:   cond.SteelCoreDia,
		AlLayers:       cond.AlLayers,
		AlAreaMM2:      cond.AlAreaMM2,
		SteelAreaMM2:   cond.SteelAreaMM2,
	}

	// Calcula elevación promedio ponderada por longitud de segmento
	var elevationAvgM float64
	if len(traceData.Coordinates) >= 2 && len(traceData.Elevations) == len(traceData.Coordinates) {
		var totalLen, weightedSum float64
		for i := 0; i < len(traceData.Coordinates)-1; i++ {
			segmentLen := geo.HaversineDistance(traceData.Coordinates[i], traceData.Coordinates[i+1])
			avgElev := (traceData.Elevations[i] + traceData.Elevations[i+1]) / 2.0
			weightedSum += avgElev * segmentLen
			totalLen += segmentLen
		}
		if totalLen > 0 {
			elevationAvgM = weightedSum / totalLen
		}
	}

	results := make([]reportResult, 0, len(data))
	tcEstAtLimit := make([]float64, 0, len(data))
	effectiveLimitForTc := cond.StaticRating
	// Usa admin max como límite efectivo si está seteado y es menor que el conductor
	if adminLimitMax > 0 && adminLimitMax < cond.StaticRating {
		effectiveLimitForTc = adminLimitMax
	} else if adminLimitMin > 0 && adminLimitMin < cond.StaticRating {
		effectiveLimitForTc = adminLimitMin
	}
	maxAbsSlope := traceData.MaxAbsSlope
	for _, point := range data {
		solPos := solar.Calculate(coord.Lat, coord.Lon, point.Time)
		w := physics.WeatherData{
			TempAir:        point.TempAir,
			WindSpeed:      point.WindSpeed,
			WindDir:        point.WindDir,
			LineAzimuth:    azimuth,
			SolarRadiation: point.SolarRadiation,
			SolarElevation: solPos.Elevation,
			PressurePa:     pressure,
		}
		std, _ := physics.GetStandard(cond.Standard)
		a := physics.Calculate(w, physCond, maxAbsSlope, std)
		tcEst := physics.EstimateConductorTemp(w, physCond, a, maxAbsSlope, std)
		tcAtLimit := physics.EstimateConductorTemp(w, physCond, effectiveLimitForTc, maxAbsSlope, std)
		results = append(results, reportResult{DataPoint: point, Ampacity: a, TcEst: tcEst, TcEstAtLimit: tcAtLimit})
		tcEstAtLimit = append(tcEstAtLimit, tcAtLimit)
	}
	doc := reportDocument{
		SchemaVersion:          "1",
		Status:                 "ok",
		CwindVersion:           version.Value,
		WeatherSource:          "Open-Meteo archive API",
		Model:                  "Balance térmico IEEE Std 738-2023",
		Conductor:              cond.Name,
		ConductorID:            id,
		Latitude:               coord.Lat,
		Longitude:              coord.Lon,
		LineAzimuth:            azimuth,
		Timezone:               timezone,
		LimiteAdminMinA:        adminLimitMin,
		LimiteAdminMaxA:        adminLimitMax,
		LimiteConductorA:       cond.StaticRating,
		RatingReferenceA:       cond.StaticRating,
		ValidationCalculatedA:  cond.CalibrationCalculatedA,
		ValidationDeviationPct: validationDeviationPct(cond),
		ElevationAvgM:          elevationAvgM,
		Results:                results,
	}
	var w io.Writer = cmd.OutOrStdout()
	var file *os.File
	if output != "" {
		file, err = os.Create(output)
		if err != nil {
			return fmt.Errorf("abrir --output: %w", err)
		}
		defer file.Close()
		w = file
	}
	return writeReport(w, format, doc, cond, showTech, false, tcEstAtLimit)
}

func chooseConductor() (conductors.Conductor, string, error) {
	list := conductors.ListMerged()
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	fmt.Fprintln(os.Stderr, styleTitle.Render("Seleccione un conductor:"))
	for i, c := range list {
		fmt.Fprintf(os.Stderr, "  %s %s — %s (%s %.0f A%s)\n",
			styleMuted.Render(fmt.Sprintf("%d)", i+1)),
			styleCyan.Render(c.ID),
			c.Name,
			styleMuted.Render(""),
			c.StaticRating,
			styleMuted.Render(""),
		)
	}
	fmt.Fprint(os.Stderr, styleMuted.Render("Opción: "))
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return conductors.Conductor{}, "", fmt.Errorf("no se pudo leer la selección")
	}
	n, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || n < 1 || n > len(list) {
		return conductors.Conductor{}, "", fmt.Errorf("selección de conductor inválida")
	}
	return list[n-1], list[n-1].ID, nil
}

func writeReport(w io.Writer, format string, doc reportDocument, cond conductors.Conductor, showStats, showTech bool, tcEstAtLimit []float64) error {
	if len(doc.Results) == 0 {
		return fmt.Errorf("no hay resultados para escribir")
	}
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(doc)
	case "csv":
		c := csv.NewWriter(w)
		metadata := [][]string{
			{"# cwind_schema_version", doc.SchemaVersion},
			{"# status", doc.Status},
			{"# cwind_version", doc.CwindVersion},
			{"# weather_source", doc.WeatherSource},
			{"# model", doc.Model},
			{"# conductor_id", doc.ConductorID},
			{"# timezone", doc.Timezone},
			{"# latitude", fmt.Sprintf("%.6f", doc.Latitude)},
			{"# longitude", fmt.Sprintf("%.6f", doc.Longitude)},
			{"# line_azimuth_deg", fmt.Sprintf("%.3f", doc.LineAzimuth)},
			{"# limite_admin_min_a", fmt.Sprintf("%.0f", doc.LimiteAdminMinA)},
			{"# limite_admin_max_a", fmt.Sprintf("%.0f", doc.LimiteAdminMaxA)},
			{"# limite_conductor_a", fmt.Sprintf("%.0f", doc.LimiteConductorA)},
			{"# rating_reference_a", fmt.Sprintf("%.1f", doc.RatingReferenceA)},
			{"# validation_calculated_a", fmt.Sprintf("%.1f", doc.ValidationCalculatedA)},
			{"# validation_deviation_pct", fmt.Sprintf("%+.2f", doc.ValidationDeviationPct)},
			{"# elevation_avg_m", fmt.Sprintf("%.1f", doc.ElevationAvgM)},
		}
		for _, row := range metadata {
			if err := c.Write(row); err != nil {
				return err
			}
		}
		if err := c.Write([]string{"time", "temperature_c", "wind_speed_mps", "wind_direction_deg", "solar_wm2", "ampacity_a", "tc_est_at_limit_c", "state", "delta_min_pct", "delta_max_pct", "delta_cond_pct"}); err != nil {
			return err
		}
		for _, r := range doc.Results {
			adminMin := doc.LimiteAdminMinA
			adminMax := doc.LimiteAdminMaxA
			conductorLimit := doc.LimiteConductorA
			state := "favorable"
			if r.Ampacity < adminMin {
				state = "crítico"
			} else if r.Ampacity < adminMax {
				state = "normal"
			} else if r.Ampacity < conductorLimit {
				state = "atención"
			}
			deltaMin := safeDeltaPct(r.Ampacity, adminMin)
			deltaMax := safeDeltaPct(r.Ampacity, adminMax)
			deltaCond := safeDeltaPct(r.Ampacity, conductorLimit)
			if err := c.Write([]string{r.Time.Format(time.RFC3339), fmt.Sprintf("%.3f", r.TempAir), fmt.Sprintf("%.3f", r.WindSpeed), fmt.Sprintf("%.3f", r.WindDir), fmt.Sprintf("%.3f", r.SolarRadiation), fmt.Sprintf("%.3f", r.Ampacity), fmt.Sprintf("%.3f", r.TcEstAtLimit), state, fmt.Sprintf("%.1f", deltaMin), fmt.Sprintf("%.1f", deltaMax), fmt.Sprintf("%.1f", deltaCond)}); err != nil {
				return err
			}
		}
		c.Flush()
		return c.Error()
	default:
		return writeTextReport(w, doc, cond, showStats, showTech, tcEstAtLimit)
	}
}

func writeTextReport(w io.Writer, doc reportDocument, cond conductors.Conductor, showStats, showTech bool, tcEstAtLimit []float64) error {
	results := doc.Results
	n := len(results)
	if n == 0 {
		return fmt.Errorf("no hay resultados para escribir")
	}

	var sum, min, max float64
	min, max = results[0].Ampacity, results[0].Ampacity

	adminMin := doc.LimiteAdminMinA
	adminMax := doc.LimiteAdminMaxA
	conductorLimit := doc.LimiteConductorA

	stateCounts := map[string]int{"crítico": 0, "normal": 0, "atención": 0, "favorable": 0}

	ampacities := make([]float64, n)
	for i, r := range results {
		sum += r.Ampacity
		ampacities[i] = r.Ampacity
		if r.Ampacity < min {
			min = r.Ampacity
		}
		if r.Ampacity > max {
			max = r.Ampacity
		}

		state := "favorable"
		if r.Ampacity < adminMin {
			state = "crítico"
		} else if r.Ampacity < adminMax {
			state = "normal"
		} else if r.Ampacity < conductorLimit {
			state = "atención"
		}
		stateCounts[state]++
	}
	avg := sum / float64(n)

	sort.Float64s(ampacities)

	tcEstAtLimitSorted := make([]float64, n)
	copy(tcEstAtLimitSorted, tcEstAtLimit)
	sort.Float64s(tcEstAtLimitSorted)
	tcMax := cond.MaxTemp

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "conductor:     %s\n", styleCyan.Render(strings.ToLower(cond.Name)))

	limitParts := []string{}
	if adminMin > 0 {
		limitParts = append(limitParts, fmt.Sprintf("mín: %.0f a", adminMin))
	}
	if adminMax > 0 {
		limitParts = append(limitParts, fmt.Sprintf("máx: %.0f a", adminMax))
	}
	limitParts = append(limitParts, fmt.Sprintf("conductor: %.0f a", conductorLimit))
	fmt.Fprintf(w, "límite: %s\n", strings.Join(limitParts, ", "))

	fromStr := strings.ToLower(results[0].Time.Format("2006-01-02 15:04"))
	toStr := strings.ToLower(results[n-1].Time.Format("2006-01-02 15:04"))
	fmt.Fprintf(w, "período:    %s a %s  (%d h)\n", fromStr, toStr, n)

	coordStr := fmt.Sprintf("%.3f %.3f", doc.Latitude, doc.Longitude)
	fmt.Fprintf(w, "traza:    %s  azimut: %.0f°\n", coordStr, doc.LineAzimuth)

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "hora  sol(w/m²)  viento(m/s)  t°amb(°c)  t°cnd(°c)  dlr(a)  Δmín%%  Δmáx%%  Δcnd%%  estado\n")
	fmt.Fprintf(w, "──────────────────────────────────────────────────────────────────────────────\n")

	for i, r := range results {
		state := "favorable"
		if r.Ampacity < adminMin {
			state = "crítico"
		} else if r.Ampacity < adminMax {
			state = "normal"
		} else if r.Ampacity < conductorLimit {
			state = "atención"
		}

		deltaCond := safeDeltaPct(r.Ampacity, conductorLimit)

		deltaMinStr := "--"
		deltaMaxStr := "--"
		deltaCondStr := fmt.Sprintf("%+6.1f%%", deltaCond)

		if adminMin > 0 {
			deltaMin := safeDeltaPct(r.Ampacity, adminMin)
			deltaMinStr = fmt.Sprintf("%+6.1f%%", deltaMin)
		} else {
			deltaMinStr = "  n/a "
		}

		if adminMax > 0 {
			deltaMax := safeDeltaPct(r.Ampacity, adminMax)
			deltaMaxStr = fmt.Sprintf("%+6.1f%%", deltaMax)
		} else {
			deltaMaxStr = "  n/a "
		}

		solStr := fmt.Sprintf("%8.0f", r.SolarRadiation)
		windStr := fmt.Sprintf("%10.1f", r.WindSpeed)
		tambStr := fmt.Sprintf("%9.1f", r.TempAir)

		tcAtLimit := tcEstAtLimit[i]
		tcStr := fmt.Sprintf("%8.0f", tcAtLimit)
		if tcAtLimit > tcMax {
			tcStr = styleAlert.Render(tcStr)
		}

		dlrStr := fmt.Sprintf("%6.0f", r.Ampacity)

		fmt.Fprintf(w, "%s  %s  %s  %s  %s  %s  %s  %s  %s  %s\n",
			styleMuted.Render(strings.ToLower(r.Time.Format("15:04"))),
			solStr,
			windStr,
			tambStr,
			tcStr,
			dlrStr,
			deltaMinStr,
			deltaMaxStr,
			deltaCondStr,
			state)
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "resumen\n")
	fmt.Fprintf(w, "crítico:   %d\n", stateCounts["crítico"])
	fmt.Fprintf(w, "normal:    %d\n", stateCounts["normal"])
	fmt.Fprintf(w, "atención:  %d\n", stateCounts["atención"])
	fmt.Fprintf(w, "favorable: %d\n", stateCounts["favorable"])

	// Calcula moda
	mode := avg
	if n > 0 {
		freq := make(map[float64]int)
		for _, a := range ampacities {
			freq[a]++
		}
		maxFreq := 0
		for val, count := range freq {
			if count > maxFreq {
				maxFreq = count
				mode = val
			}
		}
	}

	fmt.Fprintf(w, "media: %.0f a\n", avg)
	fmt.Fprintf(w, "mín:   %.0f a\n", min)
	fmt.Fprintf(w, "máx:   %.0f a\n", max)
	fmt.Fprintf(w, "moda:  %.0f a\n", mode)

	if showTech {
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "parámetros técnicos:\n")
		fmt.Fprintf(w, "  diámetro: %.4f m  │  r20: %.6f ω/m  │  α: %.4f /°c\n", cond.DiameterMeters, cond.R20, cond.Alpha)
		fmt.Fprintf(w, "  t. máx: %.0f °c  │  ε: %.2f  │  α_solar: %.2f\n", cond.MaxTemp, cond.Emissivity, cond.Absorptivity)
		if cond.SteelCoreDia > 0 {
			fmt.Fprintf(w, "  núcleo acero: %.4f m  │  al: %.0f mm²  │  acero: %.0f mm²\n", cond.SteelCoreDia, cond.AlAreaMM2, cond.SteelAreaMM2)
		}
		if cond.Source != "" {
			fmt.Fprintf(w, "  fuente: %s\n", truncate(strings.ToLower(cond.Source), 60))
		}
		if cond.RatingConditions != "" {
			fmt.Fprintf(w, "  rating: %s\n", truncate(strings.ToLower(cond.RatingConditions), 60))
		}
		if cond.CalibrationErrorPct > 0 {
			referenceRating := cond.ValidationRating()
			fmt.Fprintf(w, "  validación de referencia (no límite operativo): %.1f a vs %.0f a  │  desvío %+.2f%%  │  %s\n",
				cond.CalibrationCalculatedA, referenceRating,
				validationDeviationPct(cond), strings.ToLower(cond.CalibrationConditions))
		}
	}

	return nil
}

func validationDeviationPct(cond conductors.Conductor) float64 {
	rating := cond.ValidationRating()
	if rating <= 0 || cond.CalibrationCalculatedA <= 0 {
		return 0
	}
	return (cond.CalibrationCalculatedA/rating - 1) * 100
}

func safeDeltaPct(value, reference float64) float64 {
	if reference <= 0 {
		return 0
	}
	return (value/reference - 1) * 100
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().String("trace", "", "Archivo CSV o KML con coordenadas")
	reportCmd.Flags().String("conductor", "", "ID del conductor")
	reportCmd.Flags().String("from", "", "Fecha inicial YYYY-MM-DD")
	reportCmd.Flags().String("to", "", "Fecha final YYYY-MM-DD")
	reportCmd.Flags().Float64("line-azimuth", 0, "Azimut de la línea en grados (por defecto se calcula de la traza)")
	reportCmd.Flags().Float64("admin-min", 0, "Límite administrativo MÍN (derateo calor extremo) en Amperes (opcional, 0 para omitir)")
	reportCmd.Flags().Float64("admin-max", 0, "Límite administrativo MÁX (TC/equipamiento) en Amperes (opcional, 0 para omitir)")
	reportCmd.Flags().String("format", "text", "Formato de salida: text, json o csv")
	reportCmd.Flags().StringP("output", "o", "", "Archivo de salida (por defecto stdout)")
	reportCmd.Flags().String("timezone", "America/Argentina/Buenos_Aires", "Zona horaria IANA para fechas y weather")
	reportCmd.Flags().String("cache-dir", "", "Directorio de cache local (opt-in; vacío desactiva)")
	reportCmd.Flags().Duration("cache-ttl", 24*time.Hour, "Vigencia del cache local (0 desactiva)")
	reportCmd.Flags().String("line", "", "DEPRECADO: usá --conductor. --line se usará en el futuro para trazas geográficas.")
	reportCmd.Flags().Bool("detalle-tecnico", false, "Mostrar parámetros técnicos del conductor (R20, α, diámetro)")
	_ = reportCmd.MarkFlagRequired("trace")
	_ = reportCmd.MarkFlagRequired("from")
	_ = reportCmd.MarkFlagRequired("to")
}
