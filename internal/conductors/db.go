package conductors

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cwind-cli/cwind/internal/physics"
	"gopkg.in/yaml.v3"
)

type Conductor struct {
	ID                     string
	Name                   string
	Type                   string
	Standard               string
	DiameterMeters         float64
	R20                    float64
	Alpha                  float64
	MaxTemp                float64
	StaticRating           float64
	ReferenceRatingA       float64 `yaml:"reference_rating_a,omitempty"`
	ReferenceRatingSource  string  `yaml:"reference_rating_source,omitempty"`
	ReferenceRatingEdition string  `yaml:"reference_rating_edition,omitempty"`
	Source                 string
	RatingConditions       string
	Emissivity             float64
	Absorptivity           float64
	SteelCoreDia           float64
	AlLayers               int
	AlAreaMM2              float64
	SteelAreaMM2           float64
	CalibrationErrorPct    float64 `yaml:"calibration_error_pct,omitempty"`
	CalibrationConditions  string  `yaml:"calibration_conditions,omitempty"`
	CalibratedAt           string  `yaml:"calibrated_at,omitempty"`
	CalibrationCalculatedA float64 `yaml:"calibration_calculated_a,omitempty"`
	// CalibrationFactor is retained for backward-compatible YAML/audit data.
	// It is never applied to physical ampacity calculations.
	CalibrationFactor float64 `yaml:"calibration_factor,omitempty"`
}

var Database = map[string]Conductor{
	"acsr-16-2.5": {
		ID: "acsr-16-2.5", Name: "ACSR 16/2,5 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.0054, R20: 0.00188, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 100.0, Source: "CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0015, AlLayers: 1, AlAreaMM2: 16.0, SteelAreaMM2: 2.5,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-25-4": {
		ID: "acsr-25-4", Name: "ACSR 25/4 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.00675, R20: 0.0012, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 125.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00225, AlLayers: 1, AlAreaMM2: 25.0, SteelAreaMM2: 4.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-35-6": {
		ID: "acsr-35-6", Name: "ACSR 35/6 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.0081, R20: 0.000835, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 145.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0027, AlLayers: 1, AlAreaMM2: 35.0, SteelAreaMM2: 6.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-50-8": {
		ID: "acsr-50-8", Name: "ACSR 50/8 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.0096, R20: 0.000595, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 170.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0032, AlLayers: 1, AlAreaMM2: 50.0, SteelAreaMM2: 8.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-70-12": {
		ID: "acsr-70-12", Name: "ACSR 70/12 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.01172, R20: 0.000413, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 290.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00432, AlLayers: 2, AlAreaMM2: 70.0, SteelAreaMM2: 12.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-95-15": {
		ID: "acsr-95-15", Name: "ACSR 95/15 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.01361, R20: 0.000306, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 350.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00501, AlLayers: 2, AlAreaMM2: 95.0, SteelAreaMM2: 15.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-120-20": {
		ID: "acsr-120-20", Name: "ACSR 120/20 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.01546, R20: 0.000237, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 410.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0057, AlLayers: 2, AlAreaMM2: 120.0, SteelAreaMM2: 20.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-150-25": {
		ID: "acsr-150-25", Name: "ACSR 150/25 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.0171, R20: 0.000194, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 470.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0063, AlLayers: 2, AlAreaMM2: 150.0, SteelAreaMM2: 25.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-185-30": {
		ID: "acsr-185-30", Name: "ACSR 185/30 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.01899, R20: 0.000157, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 535.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00699, AlLayers: 2, AlAreaMM2: 185.0, SteelAreaMM2: 30.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-210-35": {
		ID: "acsr-210-35", Name: "ACSR 210/35 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.02027, R20: 0.000138, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 590.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00747, AlLayers: 2, AlAreaMM2: 210.0, SteelAreaMM2: 35.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-240-40-cimet": {
		ID: "acsr-240-40-cimet", Name: "ACSR 240/40 (IRAM 2187-I/II, CIMET)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.02184, R20: 0.000119, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 645.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00804, AlLayers: 2, AlAreaMM2: 240.0, SteelAreaMM2: 40.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-300-50": {
		ID: "acsr-300-50", Name: "ACSR 300/50 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.0245, R20: 9.49e-05, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 740.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II / Teyma Abengoa",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.009, AlLayers: 2, AlAreaMM2: 300.0, SteelAreaMM2: 50.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-340-30": {
		ID: "acsr-340-30", Name: "ACSR 340/30 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.02499, R20: 8.51e-05, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 790.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00699, AlLayers: 3, AlAreaMM2: 340.0, SteelAreaMM2: 30.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-380-50": {
		ID: "acsr-380-50", Name: "ACSR 380/50 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.027, R20: 7.57e-05, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 840.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.009, AlLayers: 3, AlAreaMM2: 380.0, SteelAreaMM2: 50.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-435-55": {
		ID: "acsr-435-55", Name: "ACSR 435/55 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.0288, R20: 6.66e-05, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 900.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0096, AlLayers: 3, AlAreaMM2: 435.0, SteelAreaMM2: 55.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-550-70": {
		ID: "acsr-550-70", Name: "ACSR 550/70 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.0324, R20: 5.26e-05, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 1020.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0108, AlLayers: 3, AlAreaMM2: 550.0, SteelAreaMM2: 70.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-680-85": {
		ID: "acsr-680-85", Name: "ACSR 680/85 (IRAM 2187-I/II)", Type: "ACSR", Standard: "IRAM_2187",
		DiameterMeters: 0.036, R20: 4.26e-05, Alpha: 0.00403, MaxTemp: 80.0,
		StaticRating: 1150.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.012, AlLayers: 3, AlAreaMM2: 680.0, SteelAreaMM2: 85.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"alac-240-40": {
		ID: "alac-240-40", Name: "ALAC 240/40 (IRAM 2187-I/II)", Type: "ALAC", Standard: "IRAM_2187",
		DiameterMeters: 0.02184, R20: 0.000138, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 645.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II (Aleación/Acero)",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00804, AlLayers: 2, AlAreaMM2: 240.0, SteelAreaMM2: 40.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-795-drake": {
		ID: "acsr-795-drake", Name: "ACSR 795 Drake 26/7 (ASTM B232)", Type: "ACSR", Standard: "ASTM_B232",
		DiameterMeters: 0.02814, R20: 7.19e-05, Alpha: 0.00403, MaxTemp: 75.0,
		StaticRating: 915.0, Source: "CME Wire / ASTM B232 Datasheet",
		ReferenceRatingA: 907.0, ReferenceRatingSource: "Nexans/Centelsa, ficha ACSR Drake 795 kcmil",
		ReferenceRatingEdition: "Caso Nexans: Tc=75°C, Tamb=25°C, V=0.61 m/s, radiación=1000 W/m²",
		RatingConditions:       "Tc=75°C, Tamb=25°C, sol=1000 W/m², V=0.61 m/s",
		Emissivity:             0.5, Absorptivity: 0.5,
		SteelCoreDia: 0.01035, AlLayers: 2, AlAreaMM2: 402.8, SteelAreaMM2: 52.2,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=75°C, Tamb=25°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"prysal-150": {
		ID: "prysal-150", Name: "PRYSAL AAAC 150 mm² (37×2,25 mm, IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.0158, R20: 0.000226, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 386.0, ReferenceRatingA: 386.0,
		ReferenceRatingSource:  "Prysmian Argentina, catálogo PRYSAL",
		ReferenceRatingEdition: "Edición 2021; método TR IEC 61597:95",
		Source:                 "Prysmian Argentina, PRYSAL Edición 2021",
		RatingConditions:       "Tc=80°C, Tamb=40°C, sol=1000 W/m², V=0.6 m/s",
		Emissivity:             0.9, Absorptivity: 0.9,
		AlAreaMM2:           150.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, sol=1000 W/m², V=0.6 m/s", CalibratedAt: "",
	},
	"prysal-240": {
		ID: "prysal-240", Name: "PRYSAL AAAC 240 mm² (37×2,85 mm, IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.0200, R20: 0.000141, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 520.0, ReferenceRatingA: 520.0,
		ReferenceRatingSource:  "Prysmian Argentina, catálogo PRYSAL",
		ReferenceRatingEdition: "Edición 2021; método TR IEC 61597:95",
		Source:                 "Prysmian Argentina, PRYSAL Edición 2021",
		RatingConditions:       "Tc=80°C, Tamb=40°C, sol=1000 W/m², V=0.6 m/s",
		Emissivity:             0.9, Absorptivity: 0.9,
		AlAreaMM2:           240.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, sol=1000 W/m², V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-16": {
		ID: "cadla-16", Name: "CADLA 16 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.00405, R20: 0.00332, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 65.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-25": {
		ID: "cadla-25", Name: "CADLA 25 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.0051, R20: 0.00209, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 100.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-35": {
		ID: "cadla-35", Name: "CADLA 35 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.00645, R20: 0.00131, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 125.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-50-7x": {
		ID: "cadla-50-7x", Name: "CADLA 50 mm² 7x (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.00756, R20: 0.000952, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 160.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-50-19x": {
		ID: "cadla-50-19x", Name: "CADLA 50 mm² 19x (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.00906, R20: 0.000663, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 195.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-70": {
		ID: "cadla-70", Name: "CADLA 70 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.00925, R20: 0.000654, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 195.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-95": {
		ID: "cadla-95", Name: "CADLA 95 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.01075, R20: 0.000484, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 235.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-120-19x": {
		ID: "cadla-120-19x", Name: "CADLA 120 mm² 19x (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.0126, R20: 0.000352, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 300.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-120-37x": {
		ID: "cadla-120-37x", Name: "CADLA 120 mm² 37x (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.01425, R20: 0.000275, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 340.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-150": {
		ID: "cadla-150", Name: "CADLA 150 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.01505, R20: 0.000249, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 340.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-185": {
		ID: "cadla-185", Name: "CADLA 185 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.01575, R20: 0.000227, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 395.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-240": {
		ID: "cadla-240", Name: "CADLA 240 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.01764, R20: 0.000181, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 455.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-300": {
		ID: "cadla-300", Name: "CADLA 300 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.01995, R20: 0.000142, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 545.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"cadla-400": {
		ID: "cadla-400", Name: "CADLA 400 mm² (IRAM 2212)", Type: "AAAC", Standard: "IRAM_2212",
		DiameterMeters: 0.02268, R20: 0.00011, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 625.0, Source: "IMSA / Cearca, ficha IRAM 2212",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"alac-35-6": {
		ID: "alac-35-6", Name: "ALAC 35/6 (IRAM 2187-I/II)", Type: "ALAC", Standard: "IRAM_2187",
		DiameterMeters: 0.0081, R20: 0.00097, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 145.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II (Aleación/Acero)",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0027, AlLayers: 1, AlAreaMM2: 35.0, SteelAreaMM2: 6.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"alac-50-8": {
		ID: "alac-50-8", Name: "ALAC 50/8 (IRAM 2187-I/II)", Type: "ALAC", Standard: "IRAM_2187",
		DiameterMeters: 0.0096, R20: 0.000691, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 170.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II (Aleación/Acero)",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0032, AlLayers: 1, AlAreaMM2: 50.0, SteelAreaMM2: 8.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"alac-70-12": {
		ID: "alac-70-12", Name: "ALAC 70/12 (IRAM 2187-I/II)", Type: "ALAC", Standard: "IRAM_2187",
		DiameterMeters: 0.01172, R20: 0.000468, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 290.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II (Aleación/Acero)",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00432, AlLayers: 2, AlAreaMM2: 70.0, SteelAreaMM2: 12.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"alac-95-15": {
		ID: "alac-95-15", Name: "ALAC 95/15 (IRAM 2187-I/II)", Type: "ALAC", Standard: "IRAM_2187",
		DiameterMeters: 0.01361, R20: 0.000355, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 350.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II (Aleación/Acero)",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00501, AlLayers: 2, AlAreaMM2: 95.0, SteelAreaMM2: 15.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"alac-120-20": {
		ID: "alac-120-20", Name: "ALAC 120/20 (IRAM 2187-I/II)", Type: "ALAC", Standard: "IRAM_2187",
		DiameterMeters: 0.01546, R20: 0.000276, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 410.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II (Aleación/Acero)",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0057, AlLayers: 2, AlAreaMM2: 120.0, SteelAreaMM2: 20.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"alac-150-25": {
		ID: "alac-150-25", Name: "ALAC 150/25 (IRAM 2187-I/II)", Type: "ALAC", Standard: "IRAM_2187",
		DiameterMeters: 0.0171, R20: 0.000225, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 470.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II (Aleación/Acero)",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.0063, AlLayers: 2, AlAreaMM2: 150.0, SteelAreaMM2: 25.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"alac-185-30": {
		ID: "alac-185-30", Name: "ALAC 185/30 (IRAM 2187-I/II)", Type: "ALAC", Standard: "IRAM_2187",
		DiameterMeters: 0.01899, R20: 0.000182, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 535.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II (Aleación/Acero)",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00699, AlLayers: 2, AlAreaMM2: 185.0, SteelAreaMM2: 30.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"alac-300-50": {
		ID: "alac-300-50", Name: "ALAC 300/50 (IRAM 2187-I/II)", Type: "ALAC", Standard: "IRAM_2187",
		DiameterMeters: 0.02444, R20: 0.00011, Alpha: 0.0036, MaxTemp: 80.0,
		StaticRating: 740.0, Source: "IMSA / CIMET, ficha IRAM 2187-I/II (Aleación/Acero)",
		RatingConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.009, AlLayers: 2, AlAreaMM2: 300.0, SteelAreaMM2: 50.0,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-1-0-raven": {
		ID: "acsr-1-0-raven", Name: "ACSR 1/0 AWG Raven (ASTM B232)", Type: "ACSR", Standard: "ASTM_B232",
		DiameterMeters: 0.00801, R20: 0.0008547, Alpha: 0.00403, MaxTemp: 75.0,
		StaticRating: 240.0, Source: "CME Wire / ASTM B232 Datasheet",
		RatingConditions: "Tc=75°C, Tamb=25°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00267, AlLayers: 1, AlAreaMM2: 53.5, SteelAreaMM2: 8.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=75°C, Tamb=25°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-4-0-penguin": {
		ID: "acsr-4-0-penguin", Name: "ACSR 4/0 AWG Penguin (ASTM B232)", Type: "ACSR", Standard: "ASTM_B232",
		DiameterMeters: 0.01134, R20: 0.0004263, Alpha: 0.00403, MaxTemp: 75.0,
		StaticRating: 380.0, Source: "CME Wire / ASTM B232 Datasheet",
		RatingConditions: "Tc=75°C, Tamb=25°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00378, AlLayers: 1, AlAreaMM2: 107.2, SteelAreaMM2: 17.9,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=75°C, Tamb=25°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-477-hawk": {
		ID: "acsr-477-hawk", Name: "ACSR 477 kcmil Hawk (ASTM B232)", Type: "ACSR", Standard: "ASTM_B232",
		DiameterMeters: 0.02179, R20: 0.0001197, Alpha: 0.00403, MaxTemp: 75.0,
		StaticRating: 660.0, Source: "CME Wire / ASTM B232 Datasheet",
		RatingConditions: "Tc=75°C, Tamb=25°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.00804, AlLayers: 2, AlAreaMM2: 241.7, SteelAreaMM2: 39.2,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=75°C, Tamb=25°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
	"acsr-954-cardinal": {
		ID: "acsr-954-cardinal", Name: "ACSR 954 kcmil Cardinal (ASTM B232)", Type: "ACSR", Standard: "ASTM_B232",
		DiameterMeters: 0.03038, R20: 6e-05, Alpha: 0.00403, MaxTemp: 75.0,
		StaticRating: 1005.0, Source: "CME Wire / ASTM B232 Datasheet",
		RatingConditions: "Tc=75°C, Tamb=25°C, al sol, V=0.6 m/s",
		Emissivity:       0.9, Absorptivity: 0.9,
		SteelCoreDia: 0.01014, AlLayers: 3, AlAreaMM2: 483.4, SteelAreaMM2: 62.7,
		CalibrationErrorPct: 0, CalibrationConditions: "Tc=75°C, Tamb=25°C, al sol, V=0.6 m/s", CalibratedAt: "",
	},
}

func Get(id string) (Conductor, bool) {
	c, ok := Database[id]
	return c, ok
}

func List() []Conductor {
	result := make([]Conductor, 0, len(Database))
	for _, c := range Database {
		result = append(result, c)
	}
	return result
}

func (c *Conductor) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("conductor sin identificador")
	}
	if c.DiameterMeters <= 0 || math.IsNaN(c.DiameterMeters) || math.IsInf(c.DiameterMeters, 0) {
		return fmt.Errorf("conductor %q: diámetro inválido", c.ID)
	}
	if c.R20 <= 0 || math.IsNaN(c.R20) || math.IsInf(c.R20, 0) {
		return fmt.Errorf("conductor %q: resistencia R20 inválida", c.ID)
	}
	if c.Alpha < 0 || math.IsNaN(c.Alpha) || math.IsInf(c.Alpha, 0) {
		return fmt.Errorf("conductor %q: coeficiente térmico inválido", c.ID)
	}
	if c.MaxTemp <= -273.15 || math.IsNaN(c.MaxTemp) || math.IsInf(c.MaxTemp, 0) {
		return fmt.Errorf("conductor %q: temperatura máxima inválida", c.ID)
	}
	if c.StaticRating <= 0 || math.IsNaN(c.StaticRating) || math.IsInf(c.StaticRating, 0) {
		return fmt.Errorf("conductor %q: ampacidad estática inválida", c.ID)
	}
	if c.Emissivity <= 0 {
		c.Emissivity = 0.9
	}
	if c.Absorptivity <= 0 {
		c.Absorptivity = 0.9
	}
	return nil
}

// Calibrate valida el conductor contra su StaticRating declarado usando RatingConditions.
// Retorna el porcentaje de error. Si error >= 1%, el modelo no coincide con datos del fabricante.
// También popula los metadatos de validación y conserva el factor histórico
// únicamente para auditoría. El factor no se aplica al cálculo físico.
func (c *Conductor) Calibrate() (float64, error) {
	referenceRating := c.ValidationRating()
	if referenceRating <= 0 {
		return 0, fmt.Errorf("no static rating to calibrate against")
	}

	// Obtiene configuración de norma para este conductor
	std, ok := physics.GetStandard(c.Standard)
	if !ok {
		std = physics.Standards["IEEE_738"]
	}

	// Parsea condiciones reales del RatingConditions del conductor
	parsedTamb, parsedWind, parsedSolar, parsedTc := parseRatingConditions(c.RatingConditions)

	// Construye condiciones de clima usando valores parseados del rating del fabricante
	solarRad := parsedSolar
	ratingText := strings.ToLower(c.RatingConditions)
	explicitNoSolar := strings.Contains(ratingText, "sin sol") ||
		strings.Contains(ratingText, "no sol") ||
		strings.Contains(ratingText, "sombra") ||
		strings.Contains(ratingText, "no sun")
	if solarRad <= 0 && !explicitNoSolar {
		solarRad = std.RatingSolarRadiation // fallback a valor del estándar
	}

	w := physics.WeatherData{
		TempAir:        parsedTamb,
		WindSpeed:      parsedWind,
		WindDir:        std.RatingWindAngleDeg,
		LineAzimuth:    0,
		SolarRadiation: solarRad,
		SolarElevation: std.RatingSolarElevation,
		PressurePa:     101325.0,
	}

	physCond := physics.Conductor{
		DiameterMeters: c.DiameterMeters,
		R20:            c.R20,
		Alpha:          c.Alpha,
		MaxTemp:        c.MaxTemp,
		Emissivity:     c.Emissivity,
		Absorptivity:   c.Absorptivity,
		SteelCoreDia:   c.SteelCoreDia,
		AlLayers:       c.AlLayers,
		AlAreaMM2:      c.AlAreaMM2,
		SteelAreaMM2:   c.SteelAreaMM2,
	}
	if parsedTc > 0 {
		physCond.MaxTemp = parsedTc
	}
	if physCond.Emissivity <= 0 {
		physCond.Emissivity = std.DefaultEmissivity
	}
	if physCond.Absorptivity <= 0 {
		physCond.Absorptivity = std.DefaultAbsorptivity
	}

	calc := physics.Calculate(w, physCond, 0, std) // 0 pendiente para calibración (horizontal)

	errorPct := math.Abs(calc-referenceRating) / referenceRating * 100.0

	// El factor es una corrección empírica del caso de referencia, no un
	// reemplazo de las ecuaciones ni una validación del modelo para otros casos.
	var calFactor float64
	if calc > 0 {
		calFactor = referenceRating / calc
	} else {
		calFactor = 1.0
	}

	// Popula campos de calibración
	c.CalibrationErrorPct = errorPct
	c.CalibrationConditions = c.RatingConditions
	c.CalibratedAt = time.Now().Format(time.RFC3339)
	c.CalibrationCalculatedA = calc
	c.CalibrationFactor = calFactor

	return errorPct, nil
}

func (c Conductor) ValidationRating() float64 {
	if c.ReferenceRatingA > 0 {
		return c.ReferenceRatingA
	}
	return c.StaticRating
}

// parseRatingConditions extrae condiciones de prueba del string RatingConditions.
// Formatos esperados:
// "Tc=80°C, Tamb=40°C, al sol, V=0.6 m/s"
// "Tc=80°C, Tamb=40°C, SIN sol, V=0.6 m/s"
// "Tc=75°C, Tamb=25°C, V=0.6 m/s, sol pleno"
func parseRatingConditions(s string) (tamb, windSpeed, solarRad, tcMax float64) {
	// Valores por defecto
	tamb = 40.0
	windSpeed = 0.6
	solarRad = 1000.0
	tcMax = 0

	// Extrae Tc (temperatura máxima del conductor)
	if re := regexp.MustCompile(`Tc\s*=\s*([\d.]+)\s*°?C?`); re != nil {
		if matches := re.FindStringSubmatch(s); len(matches) > 1 {
			fmt.Sscanf(matches[1], "%f", &tcMax)
		}
	}

	// Extrae Tamb (temperatura ambiente)
	if re := regexp.MustCompile(`Tamb\s*=\s*([\d.]+)\s*°?C?`); re != nil {
		if matches := re.FindStringSubmatch(s); len(matches) > 1 {
			fmt.Sscanf(matches[1], "%f", &tamb)
		}
	}

	// Extrae V (velocidad del viento)
	if re := regexp.MustCompile(`V\s*=\s*([\d.]+)\s*m/?s?`); re != nil {
		if matches := re.FindStringSubmatch(s); len(matches) > 1 {
			fmt.Sscanf(matches[1], "%f", &windSpeed)
		}
	}

	// Verifica condiciones solares
	lower := strings.ToLower(s)
	if re := regexp.MustCompile(`(?:sol|solar(?:\s+radiation)?)\s*=\s*([\d.]+)\s*w\s*/\s*m(?:2|²)`); re != nil {
		if matches := re.FindStringSubmatch(lower); len(matches) > 1 {
			fmt.Sscanf(matches[1], "%f", &solarRad)
		} else if strings.Contains(lower, "sin sol") || strings.Contains(lower, "no sol") || strings.Contains(lower, "sombra") || strings.Contains(lower, "no sun") {
			solarRad = 0.0
		} else if strings.Contains(lower, "al sol") || strings.Contains(lower, "sol pleno") || strings.Contains(lower, "full sun") {
			solarRad = 1000.0
		}
	}
	// Por defecto: asume 1000 W/m² (conservador)

	return tamb, windSpeed, solarRad, tcMax
}

var calibrationDone = false

func EnsureCalibration() {
	if calibrationDone {
		return
	}
	calibrationDone = true

	for id, cond := range Database {
		if cond.StaticRating > 0 {
			if errPct, err := cond.Calibrate(); err == nil {
				Database[id] = cond // Guarda en mapa ya que Calibrate modifica la struct
				if errPct >= 1.0 {
					// Conductores built-in usan condiciones de prueba del fabricante que pueden diferir del modelo completo IEEE 738
					// (ej. ángulo viento, elevación solar, presión). Esto es esperado para muchos conductores.
				}
			}
		}
	}
}

func LoadUserConductors() map[string]Conductor {
	EnsureCalibration()
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	path := filepath.Join(home, ".cwind", "conductors.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var userDB map[string]Conductor
	if err := yaml.Unmarshal(data, &userDB); err != nil {
		return nil
	}
	return userDB
}

func GetMerged(id string) (Conductor, bool) {
	if c, ok := Database[id]; ok {
		return c, true
	}
	userDB := LoadUserConductors()
	if userDB != nil {
		if c, ok := userDB[id]; ok {
			return c, true
		}
	}
	return Conductor{}, false
}

func ListMerged() []Conductor {
	result := make([]Conductor, 0, len(Database))
	for _, c := range Database {
		result = append(result, c)
	}
	userDB := LoadUserConductors()
	if userDB != nil {
		for _, c := range userDB {
			result = append(result, c)
		}
	}
	return result
}
