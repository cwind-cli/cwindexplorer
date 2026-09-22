package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/cwind-cli/cwind/internal/conductors"
	"github.com/cwind-cli/cwind/internal/elevation"
	"github.com/cwind-cli/cwind/internal/geo"
	"github.com/cwind-cli/cwind/internal/physics"
	"github.com/cwind-cli/cwind/internal/solar"
	"github.com/cwind-cli/cwind/internal/version"
	"github.com/cwind-cli/cwind/internal/weather"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

var (
	styleBannerTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d47e")).Bold(true)
	styleBannerSub   = lipgloss.NewStyle().Foreground(lipgloss.Color("#b794f4"))
	styleBannerMuted = lipgloss.NewStyle().Foreground(lipgloss.Color("#5c6a7e"))
	stylePrompt      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d4d4")).Bold(true)
	styleAddMuted    = lipgloss.NewStyle().Foreground(lipgloss.Color("#5c6a7e"))
	styleWarning     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb020"))
	styleAddSub      = lipgloss.NewStyle().Foreground(lipgloss.Color("#b794f4"))
	styleAddSuccess  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d47e")).Bold(true)
	styleAddInfo     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d4d4"))
	styleAlert       = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b35")).Bold(true)
	styleNormal      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d47e"))
	styleOrange      = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff9f1c"))
)

var rootCmd = &cobra.Command{
	Use:     "cwind",
	Short:   "Cwind - Reportes de ampacidad térmica dinámica (DLR)",
	Long:    `Herramienta para medir la ampacidad de conductores de alta tensión.`,
	RunE:    runInteractive,
	Version: version.Value,
}

func init() {
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {})
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", cmd.Long)
		fmt.Fprintf(cmd.OutOrStdout(), "Usage:\n  %s\n\n", cmd.UseLine())
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name != "help" {
				fmt.Fprintf(cmd.OutOrStdout(), "  --%-20s %s\n", f.Name, f.Usage)
			}
		})
		fmt.Fprintf(cmd.OutOrStdout(), "\nAvailable Commands:\n")
		for _, c := range cmd.Commands() {
			if c.Hidden == false {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", c.Use, c.Short)
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\nFlags:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  -h, --help      help for cwind\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  -v, --version   version for cwind\n\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Use \"cwind [command] --help\" for more information about a command.\n")
	})

	rootCmd.AddCommand(restartCmd)
}

func runInteractive(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return nil
	}

	showBannerIfFirstRun()

	for {
		sector, err := askSector()
		if err != nil {
			return err
		}

		if len(sector.Submodulos) > 0 {
			submodulo, err := askSubmodulo(sector)
			if err != nil {
				return err
			}
			if submodulo == nil {
				// Usuario volvió o no hay módulos disponibles
				continue
			}
			if err := submodulo.Ejecutar(cmd); err != nil {
				if err.Error() == "volver al menú anterior" {
					continue
				}
				return err
			}
			// Módulo funcional completado exitosamente - salir
			return nil
		} else {
			if err := sector.Ejecutar(cmd); err != nil {
				return err
			}
			// Sector sin submódulos completado - salir
			return nil
		}
	}
}

func askSector() (Sector, error) {
	sectores := GetSectores()
	if len(sectores) == 0 {
		return Sector{}, fmt.Errorf("no hay sectores registrados")
	}

	options := make([]huh.Option[string], len(sectores))
	for i, s := range sectores {
		options[i] = huh.NewOption(s.Nombre(), s.Nombre())
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("¿Qué sector querés analizar?").
				Options(options...).
				Value(&selected),
		),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
	if err := form.Run(); err != nil {
		return Sector{}, fmt.Errorf("cancelado: %w", err)
	}

	sector := FindSector(selected)
	if sector == nil {
		return Sector{}, fmt.Errorf("sector no encontrado: %s", selected)
	}
	return *sector, nil
}

func askSubmodulo(sector Sector) (Modulo, error) {
	submodulos := sector.Submodulos
	if len(submodulos) == 0 {
		return nil, nil
	}

	available := make([]Modulo, 0, len(submodulos))
	unavailable := make([]string, 0)
	for _, m := range submodulos {
		if m.Disponible() {
			available = append(available, m)
		} else {
			unavailable = append(unavailable, styleAddMuted.Render("  "+m.Nombre()))
		}
	}
	options := make([]huh.Option[string], len(available)+1)
	for i, m := range available {
		options[i] = huh.NewOption(m.Nombre(), m.Nombre())
	}
	options[len(available)] = huh.NewOption("← Volver", "back")

	var selected string
	fields := []huh.Field{
		huh.NewSelect[string]().
			Title(fmt.Sprintf("%s:", sector.Nombre())).
			Options(options...).
			Value(&selected),
	}
	if len(unavailable) > 0 {
		fields = append(fields, huh.NewNote().
			Title(styleAddMuted.Render(strings.Join(unavailable, "\n"))).
			Description(styleAddMuted.Render("No disponibles")))
	}
	form := huh.NewForm(
		huh.NewGroup(fields...),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("cancelado: %w", err)
	}

	if selected == "back" {
		return nil, nil
	}

	for _, m := range available {
		if m.Nombre() == selected {
			return m, nil
		}
	}
	return nil, fmt.Errorf("submódulo no encontrado: %s", selected)
}

func showBannerIfFirstRun() {
	printBanner()
}

func printBanner() {
	banner := `

     ██████╗██╗    ██╗██╗███╗   ██╗██████╗
    ██╔════╝██║    ██║██║████╗  ██║██╔══██╗
    ██║     ██║ █╗ ██║██║██╔██╗ ██║██║  ██║
    ██║     ██║███╗██║██║██║╚██╗██║██║  ██║
    ╚██████╗╚███╔███╔╝██║██║ ╚████║██████╔╝
     ╚═════╝ ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═════╝

`
	fmt.Println(styleBannerTitle.Render(banner))
	fmt.Println(styleBannerTitle.Render("                   v" + version.Value))
	fmt.Println()
}

func askTraceFile() (string, error) {
	var tracePath string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Archivo de traza (CSV/KML)").
				Placeholder("ej: linea.csv").
				Value(&tracePath).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("requerido")
					}
					if _, err := os.Stat(s); os.IsNotExist(err) {
						return fmt.Errorf("archivo no encontrado: %s", s)
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
	if err := form.Run(); err != nil {
		return "", fmt.Errorf("cancelado: %w", err)
	}
	return tracePath, nil
}

func askConductor() (conductors.Conductor, string, error) {
	list := conductors.ListMerged()
	if len(list) == 0 {
		return conductors.Conductor{}, "", fmt.Errorf("no hay conductores disponibles")
	}

	options := make([]huh.Option[string], len(list)+1)
	for i, c := range list {
		options[i] = huh.NewOption(fmt.Sprintf("%s — %s (%.0f A)", c.ID, c.Name, c.StaticRating), c.ID)
	}

	options[len(list)] = huh.NewOption("← Volver", "back")

	var selectedID string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Seleccionar conductor (Esc para volver)").
				Options(options...).
				Value(&selectedID),
		),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
	if err := form.Run(); err != nil {
		return conductors.Conductor{}, "", fmt.Errorf("cancelado: %w", err)
	}

	if selectedID == "back" {
		return askConductor() // Vuelve al menú de elección
	}

	conductors.EnsureCalibration()
	cond, ok := conductors.GetMerged(selectedID)
	if !ok {
		return conductors.Conductor{}, "", fmt.Errorf("conductor no encontrado: %s", selectedID)
	}
	return cond, selectedID, nil
}

var errBack = fmt.Errorf("back")

func runConductorAddWizard() (conductors.Conductor, string, error) {
	// Ejecuta el wizard de agregar conductor en línea
	var (
		name             string
		conductorType    string
		diameterMMStr    string
		r20PerKmStr      string
		maxTempStr       string
		staticRatingStr  string
		source           string
		ratingConditions string
		emissivityStr    string
		absorptivityStr  string
		steelCoreDiaStr  string
		alAreaStr        string
		steelAreaStr     string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Nombre comercial *").
				Placeholder("ej: ACSR 240/40 MiEmpresa").
				Value(&name).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("requerido")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Tipo de conductor * (Esc para volver)").
				Options(
					huh.NewOption("ACSR (aluminio-acero)", "ACSR"),
					huh.NewOption("AAAC (aleación aluminio)", "AAAC"),
					huh.NewOption("ACAR (aluminio refuerzo aluminio)", "ACAR"),
					huh.NewOption("AAC (aluminio puro)", "AAC"),
					huh.NewOption("Otro", "OTRO"),
					huh.NewOption("← Volver", "back"),
				).
				Value(&conductorType).
				Validate(func(s string) error {
					if s == "back" {
						return errBack
					}
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("requerido")
					}
					return nil
				}),
			huh.NewInput().
				Title("Diámetro exterior (mm) *").
				Placeholder("ej: 21.84").
				Value(&diameterMMStr).
				Validate(func(s string) error {
					var v float64
					if _, err := fmt.Sscanf(s, "%f", &v); err != nil || v <= 0 {
						return fmt.Errorf("debe ser un número positivo")
					}
					return nil
				}),
			huh.NewInput().
				Title("Resistencia a 20°C (Ω/km) *").
				Placeholder("ej: 0.119").
				Value(&r20PerKmStr).
				Validate(func(s string) error {
					var v float64
					if _, err := fmt.Sscanf(s, "%f", &v); err != nil || v <= 0 {
						return fmt.Errorf("debe ser un número positivo")
					}
					return nil
				}),
			huh.NewInput().
				Title("Temperatura máxima de diseño (°C) *").
				Placeholder("80").
				Value(&maxTempStr).
				Validate(func(s string) error {
					var v float64
					if _, err := fmt.Sscanf(s, "%f", &v); err != nil || v <= -273.15 || v > 250 {
						return fmt.Errorf("inválida (rango típico: 50-250°C)")
					}
					return nil
				}),
			huh.NewInput().
				Title("Rating de catálogo (A) - opcional").
				Placeholder("ej: 645 (vacío para omitir)").
				Value(&staticRatingStr),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Emisividad (ε) - opcional").
				Placeholder("0.9 (Enter para default)").
				Value(&emissivityStr),
			huh.NewInput().
				Title("Absortividad solar (α) - opcional").
				Placeholder("0.9 (Enter para default)").
				Value(&absorptivityStr),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Diámetro núcleo de acero (mm) - solo ACSR").
				Placeholder("ej: 7.2 (vacío para omitir)").
				Value(&steelCoreDiaStr),
			huh.NewInput().
				Title("Área aluminio (mm²) - solo ACSR").
				Placeholder("ej: 240 (vacío para omitir)").
				Value(&alAreaStr),
			huh.NewInput().
				Title("Área acero (mm²) - solo ACSR").
				Placeholder("ej: 40 (vacío para omitir)").
				Value(&steelAreaStr),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Fuente/procedencia del dato *").
				Placeholder("ej: ficha fabricante X, IRAM 2187").
				Value(&source).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("requerido")
					}
					return nil
				}),
			huh.NewInput().
				Title("Condiciones de rating *").
				Placeholder("ej: Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s").
				Value(&ratingConditions).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("requerido")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())

	if err := form.Run(); err != nil {
		// Usuario presionó Esc o canceló - vuelve
		return conductors.Conductor{}, "", errBack
	}

	var diameterMM, r20PerKm, maxTemp, staticRating float64
	fmt.Sscanf(diameterMMStr, "%f", &diameterMM)
	fmt.Sscanf(r20PerKmStr, "%f", &r20PerKm)
	fmt.Sscanf(maxTempStr, "%f", &maxTemp)
	fmt.Sscanf(staticRatingStr, "%f", &staticRating)

	emissivity := 0.9
	if strings.TrimSpace(emissivityStr) != "" {
		fmt.Sscanf(emissivityStr, "%f", &emissivity)
	}
	absorptivity := 0.9
	if strings.TrimSpace(absorptivityStr) != "" {
		fmt.Sscanf(absorptivityStr, "%f", &absorptivity)
	}
	steelCoreDia := 0.0
	if strings.TrimSpace(steelCoreDiaStr) != "" {
		var v float64
		fmt.Sscanf(steelCoreDiaStr, "%f", &v)
		steelCoreDia = v / 1000.0
	}
	alAreaMM2 := 0.0
	if strings.TrimSpace(alAreaStr) != "" {
		fmt.Sscanf(alAreaStr, "%f", &alAreaMM2)
	}
	steelAreaMM2 := 0.0
	if strings.TrimSpace(steelAreaStr) != "" {
		fmt.Sscanf(steelAreaStr, "%f", &steelAreaMM2)
	}

	id := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	id = strings.ReplaceAll(id, "/", "-")
	id = strings.ReplaceAll(id, ".", "-")

	cond := conductors.Conductor{
		ID:               id,
		Name:             name,
		Type:             conductorType,
		DiameterMeters:   diameterMM / 1000.0,
		R20:              r20PerKm / 1000.0,
		Alpha:            0.00403,
		MaxTemp:          maxTemp,
		StaticRating:     staticRating,
		Source:           source,
		RatingConditions: ratingConditions,
		Emissivity:       emissivity,
		Absorptivity:     absorptivity,
		SteelCoreDia:     steelCoreDia,
		AlAreaMM2:        alAreaMM2,
		SteelAreaMM2:     steelAreaMM2,
	}

	if err := cond.Validate(); err != nil {
		return conductors.Conductor{}, "", fmt.Errorf("conductor inválido: %w", err)
	}

	// Ejecuta calibración si se proveyó rating estático
	if staticRating > 0 {
		errPct, _ := cond.Calibrate()
		cond.CalibrationErrorPct = errPct
		cond.CalibrationConditions = ratingConditions
		cond.CalibratedAt = time.Now().Format(time.RFC3339)

		if errPct >= 1.0 {
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("═══════════════════════════════════════════════════════════════════\n"))
			fmt.Fprintf(os.Stderr, styleAlert.Render("⚠  CALIBRACIÓN FALLÓ: Error %.2f%% ≥ 1%%\n"), errPct)
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("────────────────────────────────────────────────────────────────────\n"))
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("El balance IEEE 738-2023 no reproduce el rating de fabricante declarado bajo las condiciones disponibles.\n"))
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("Rating calculado: %.0f A  |  Rating fabricante: %.0f A\n"),
				cond.StaticRating*(1-errPct/100), cond.StaticRating)
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("La corrección solo está validada en las condiciones de referencia; fuera de ellas aumenta la incertidumbre.\n"))
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("\n"))
			fmt.Fprintf(os.Stderr, styleAddSub.Render("Opciones:\n"))
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("  1) Continuar y guardar; el rating quedará como referencia de validación\n"))
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("  2) Cancelar y elegir conductor existente\n"))
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("  3) Cancelar y revisar parámetros\n"))
			fmt.Fprintf(os.Stderr, styleAddMuted.Render("────────────────────────────────────────────────────────────────────\n"))

			var choice string
			choiceForm := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("¿Cómo desea proceder?").
						Options(
							huh.NewOption("Continuar (guardar con advertencia)", "continue"),
							huh.NewOption("Cancelar y elegir conductor existente", "choose_existing"),
							huh.NewOption("Cancelar y revisar parámetros", "retry"),
						).
						Value(&choice),
				),
			).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())

			if err := choiceForm.Run(); err != nil {
				return conductors.Conductor{}, "", fmt.Errorf("cancelado: %w", err)
			}

			switch choice {
			case "choose_existing":
				return askConductor() // Llamada recursiva para seleccionar de BD
			case "retry":
				return runConductorAddWizard() // Llamada recursiva para reintentar
			case "continue":
				// Continúa con advertencia
				fmt.Fprintf(os.Stderr, styleWarning.Render("⚠ Continuando con calibración fuera de tolerancia (%.2f%%)\n"), errPct)
			}
		} else if errPct > 0 {
			fmt.Fprintf(os.Stderr, styleNormal.Render("✓ Calibración exitosa: error %.2f%% (< 1%%)\n"), errPct)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return conductors.Conductor{}, "", fmt.Errorf("no se pudo obtener directorio home: %w", err)
	}
	cwindDir := filepath.Join(home, ".cwind")
	if err := os.MkdirAll(cwindDir, 0755); err != nil {
		return conductors.Conductor{}, "", fmt.Errorf("no se pudo crear %s: %w", cwindDir, err)
	}

	yamlPath := filepath.Join(cwindDir, "conductors.yaml")

	var userDB map[string]conductors.Conductor
	if data, err := os.ReadFile(yamlPath); err == nil {
		_ = yaml.Unmarshal(data, &userDB)
	}
	if userDB == nil {
		userDB = make(map[string]conductors.Conductor)
	}

	if _, exists := userDB[id]; exists {
		fmt.Fprintf(os.Stderr, styleAddMuted.Render("aviso: conductor %q ya existe en %s, se sobrescribirá\n"), id, yamlPath)
	}
	userDB[id] = cond

	out, err := yaml.Marshal(userDB)
	if err != nil {
		return conductors.Conductor{}, "", fmt.Errorf("error serializando YAML: %w", err)
	}
	if err := os.WriteFile(yamlPath, out, 0644); err != nil {
		return conductors.Conductor{}, "", fmt.Errorf("error guardando %s: %w", yamlPath, err)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, styleAddSuccess.Render("✓ Conductor guardado correctamente"))
	fmt.Fprintln(os.Stderr, styleAddMuted.Render("────────────────────────────────────────"))
	fmt.Fprintf(os.Stderr, "  %s %s\n", styleAddInfo.Render("Archivo:"), yamlPath)
	fmt.Fprintf(os.Stderr, "  %s %s\n", styleAddInfo.Render("ID:"), styleAddSub.Render(id))
	fmt.Fprintf(os.Stderr, "  %s %s\n", styleAddInfo.Render("Nombre:"), name)
	fmt.Fprintf(os.Stderr, "  %s %s\n", styleAddInfo.Render("Tipo:"), conductorType)
	fmt.Fprintf(os.Stderr, "  %s %.2f mm\n", styleAddInfo.Render("Diámetro:"), diameterMM)
	fmt.Fprintf(os.Stderr, "  %s %.6f Ω/km\n", styleAddInfo.Render("R20:"), r20PerKm)
	fmt.Fprintf(os.Stderr, "  %s %.0f °C\n", styleAddInfo.Render("Temp. máx:"), maxTemp)
	if staticRating > 0 {
		fmt.Fprintf(os.Stderr, "  %s %.0f A\n", styleAddInfo.Render("Rating:"), staticRating)
	}
	if steelCoreDia > 0 {
		fmt.Fprintf(os.Stderr, "  %s %.2f mm\n", styleAddInfo.Render("Núcleo acero:"), steelCoreDia*1000)
	}
	if alAreaMM2 > 0 {
		fmt.Fprintf(os.Stderr, "  %s %.0f mm²\n", styleAddInfo.Render("Área Al:"), alAreaMM2)
	}
	if steelAreaMM2 > 0 {
		fmt.Fprintf(os.Stderr, "  %s %.0f mm²\n", styleAddInfo.Render("Área Acero:"), steelAreaMM2)
	}
	fmt.Fprintf(os.Stderr, "  %s ε=%.2f  α=%.2f\n", styleAddInfo.Render("Superficie:"), emissivity, absorptivity)
	fmt.Fprintf(os.Stderr, "  %s %s\n", styleAddInfo.Render("Fuente:"), source)
	fmt.Fprintf(os.Stderr, "  %s %s\n", styleAddInfo.Render("Condiciones:"), ratingConditions)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, styleAddMuted.Render("Usalo con: cwind report --conductor "+id+" --trace linea.csv --from 2024-01-01 --to 2024-01-02"))

	return cond, id, nil
}

func askAdminLimit() (float64, float64, error) {
	var adminMinStr, adminMaxStr string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Límite administrativo MÍN (derateo calor extremo) - opcional (Enter para omitir)").
				Placeholder("Ej: 580").
				Value(&adminMinStr),
			huh.NewInput().
				Title("Límite administrativo MÁX (TC/equipamiento) - opcional (Enter para omitir)").
				Placeholder("Ej: 600").
				Value(&adminMaxStr),
		),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
	if err := form.Run(); err != nil {
		return 0, 0, fmt.Errorf("cancelado: %w", err)
	}
	var adminMin, adminMax float64
	if strings.TrimSpace(adminMinStr) != "" {
		v, err := strconv.ParseFloat(strings.TrimSpace(adminMinStr), 64)
		if err != nil {
			return 0, 0, fmt.Errorf("valor inválido para límite admin mín: %w", err)
		}
		adminMin = v
	}
	if strings.TrimSpace(adminMaxStr) != "" {
		v, err := strconv.ParseFloat(strings.TrimSpace(adminMaxStr), 64)
		if err != nil {
			return 0, 0, fmt.Errorf("valor inválido para límite admin máx: %w", err)
		}
		adminMax = v
	}
	return adminMin, adminMax, nil
}

func askDateRange() (time.Time, time.Time, error) {
	var rangeChoice string
	now := time.Now()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Rango de fechas").
				Options(
					huh.NewOption("Última 1 hora", "1h"),
					huh.NewOption("Últimas 12 horas", "12h"),
					huh.NewOption("Últimas 24 horas", "24h"),
					huh.NewOption("Elegir fecha...", "custom"),
				).
				Value(&rangeChoice),
		),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
	if err := form.Run(); err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("cancelado: %w", err)
	}

	switch rangeChoice {
	case "1h":
		return now.Add(-1 * time.Hour), now, nil
	case "12h":
		return now.Add(-12 * time.Hour), now, nil
	case "24h":
		return now.Add(-24 * time.Hour), now, nil
	case "custom":
		return askCustomDates()
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("opción inválida")
	}
}

func askCustomDates() (time.Time, time.Time, error) {
	var fromStr, toStr string
	location, _ := time.LoadLocation("America/Argentina/Buenos_Aires")

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Fecha y hora inicial (YYYY-MM-DD HH:MM)").
				Placeholder("ej: 2024-01-01 08:00").
				Value(&fromStr).
				Validate(func(s string) error {
					_, err := time.ParseInLocation("2006-01-02 15:04", s, location)
					return err
				}),
			huh.NewInput().
				Title("Fecha y hora final (YYYY-MM-DD HH:MM)").
				Placeholder("ej: 2024-01-01 20:00").
				Value(&toStr).
				Validate(func(s string) error {
					_, err := time.ParseInLocation("2006-01-02 15:04", s, location)
					return err
				}),
		),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
	if err := form.Run(); err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("cancelado: %w", err)
	}

	from, _ := time.ParseInLocation("2006-01-02 15:04", fromStr, location)
	to, _ := time.ParseInLocation("2006-01-02 15:04", toStr, location)

	if to.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("--to debe ser posterior o igual a --from")
	}
	if to.Sub(from).Hours() > 24 {
		return time.Time{}, time.Time{}, fmt.Errorf("rango máximo es 24h")
	}
	return from, to, nil
}
func runReportDirect(cmd *cobra.Command, tracePath string, cond conductors.Conductor, condID string, adminMin, adminMax float64, from, to time.Time) error {
	timezone := "America/Argentina/Buenos_Aires"

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
	azimuth := traceData.Azimuth()

	data, err := weather.FetchRangeContextWithCache(cmd.Context(), coord.Lat, coord.Lon, from, to, timezone, "", 24*time.Hour)
	if err != nil {
		return fmt.Errorf("weather: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("la respuesta meteorológica no contiene datos")
	}

	filtered := make([]weather.DataPoint, 0, len(data))
	for _, d := range data {
		if !d.Time.Before(from) && !d.Time.After(to) {
			filtered = append(filtered, d)
		}
	}
	if len(filtered) == 0 {
		return fmt.Errorf("no hay datos meteorológicos en el rango solicitado")
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

	elevationAvgM := 0.0

	maxAbsSlope := traceData.MaxAbsSlope
	results := make([]reportResult, 0, len(filtered))
	tcEstAtLimit := make([]float64, 0, len(filtered))

	// Usa admin max como límite efectivo si está seteado y es menor que el conductor
	// Si admin max no está seteado pero admin min sí, usa admin min
	effectiveLimitForTc := cond.StaticRating
	if adminMax > 0 && adminMax < cond.StaticRating {
		effectiveLimitForTc = adminMax
	} else if adminMin > 0 && adminMin < cond.StaticRating {
		effectiveLimitForTc = adminMin
	}

	for _, point := range filtered {
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
		SchemaVersion:    "1",
		Status:           "ok",
		CwindVersion:     version.Value,
		WeatherSource:    "Open-Meteo archive API",
		Model:            "Balance térmico IEEE Std 738-2023",
		Conductor:        cond.Name,
		ConductorID:      condID,
		Latitude:         coord.Lat,
		Longitude:        coord.Lon,
		LineAzimuth:      azimuth,
		Timezone:         timezone,
		LimiteAdminMinA:  adminMin,
		LimiteAdminMaxA:  adminMax,
		LimiteConductorA: cond.StaticRating,
		ElevationAvgM:    elevationAvgM,
		Results:          results,
	}

	// En terminales interactivas se presenta como dashboard; stdout no-TTY
	// conserva el reporte plano para pipes, CI y compatibilidad.
	action, err := writeInteractiveDashboard(cmd, doc, cond, tcEstAtLimit, tracePath, adminMin, adminMax)
	if err != nil {
		return err
	}

	if action == "exportar" {
		return runExportWizard(cmd, doc, cond)
	}
	if action == "finalizar" {
		return nil
	}

	// Compatibilidad para cualquier implementación de dashboard que no elija acción.
	return runAnalysisLoop(cmd, doc, cond, adminMin, adminMax, results, tcEstAtLimit)
}

func writeInteractiveDashboard(cmd *cobra.Command, doc reportDocument, cond conductors.Conductor, tcEstAtLimit []float64, tracePath string, adminMin, adminMax float64) (string, error) {
	var report bytes.Buffer
	if err := writeReport(&report, "text", doc, cond, false, false, tcEstAtLimit); err != nil {
		return "", err
	}

	return runDLRTUI(cmd.OutOrStdout(), report.String(), cond, tracePath, adminMin, adminMax)
}

func terminalWidth(file *os.File) int {
	width, _, err := term.GetSize(int(file.Fd()))
	if err != nil || width <= 0 {
		return 120
	}
	return width
}

func runAnalysisLoop(cmd *cobra.Command, doc reportDocument, cond conductors.Conductor, adminMin, adminMax float64, results []reportResult, tcEstAtLimit []float64) error {
	for {
		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("ACCIONES").
					Options(
						huh.NewOption("Exportar reporte (CSV/JSON)", "exportar"),
						huh.NewOption("Finalizar sesión", "finalizar"),
					).
					Value(&choice),
			),
		).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
		if err := form.Run(); err != nil {
			return nil
		}

		switch choice {
		case "exportar":
			err := runExportWizard(cmd, doc, cond)
			if err != nil {
				return err
			}
			// Exit after successful export
			return nil
		case "finalizar":
			return nil
		}
	}
}

func runExportWizard(cmd *cobra.Command, doc reportDocument, cond conductors.Conductor) error {
	var formatChoice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("formato de exportación").
				Options(
					huh.NewOption("csv", "csv"),
					huh.NewOption("json", "json"),
				).
				Value(&formatChoice),
		),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
	if err := form.Run(); err != nil {
		return nil
	}

	var outputPath string
	inputForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("archivo de salida").
				Placeholder(fmt.Sprintf("reporte.%s", formatChoice)).
				Value(&outputPath),
		),
	).WithTheme(huh.ThemeBase()).WithWidth(wizardWidth())
	if err := inputForm.Run(); err != nil {
		return nil
	}

	if strings.TrimSpace(outputPath) == "" {
		outputPath = fmt.Sprintf("reporte.%s", formatChoice)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("crear archivo: %w", err)
	}
	defer f.Close()

	err = writeReport(f, formatChoice, doc, cond, false, false, nil)
	if err != nil {
		return fmt.Errorf("escribir reporte: %w", err)
	}

	fmt.Printf("\nexportado a: %s\n", outputPath)
	return nil
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Reinicia el wizard interactivo",
	RunE:  runInteractive,
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		return err
	}
	return nil
}

func wizardWidth() int {
	if file, ok := os.Stdout.Stat(); ok == nil && file.Mode()&os.ModeCharDevice != 0 {
		width := terminalWidth(os.Stdout)
		if width >= 120 {
			return 72
		}
		if width >= 90 {
			return 64
		}
	}
	return 56
}
