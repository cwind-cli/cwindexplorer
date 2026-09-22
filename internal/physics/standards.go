package physics

type StandardConfig struct {
	Name                 string
	RatingWindAngleDeg   float64
	RatingSolarElevation float64
	RatingSolarRadiation float64
	DefaultEmissivity    float64
	DefaultAbsorptivity  float64
	ACResistanceMethod   ACResistanceMethod
	ConvectionMethod     ConvectionMethod
}

type ACResistanceMethod int

const (
	ACResistance_IEEE738 ACResistanceMethod = iota
	ACResistance_IRAM2187
	ACResistance_IRAM2212
	ACResistance_ASTMB232
)

type ConvectionMethod int

const (
	Convection_IEEE738 ConvectionMethod = iota
	Convection_IRAM2187
)

var Standards = map[string]StandardConfig{
	"IRAM_2187": {
		Name:                 "IRAM 2187-I/II",
		RatingWindAngleDeg:   90.0,
		RatingSolarElevation: 80.0,
		RatingSolarRadiation: 1000.0,
		DefaultEmissivity:    0.5,
		DefaultAbsorptivity:  0.5,
		ACResistanceMethod:   ACResistance_IRAM2187,
		ConvectionMethod:     Convection_IRAM2187,
	},
	"IRAM_2212": {
		Name:                 "IRAM 2212",
		RatingWindAngleDeg:   90.0,
		RatingSolarElevation: 80.0,
		RatingSolarRadiation: 1000.0,
		DefaultEmissivity:    0.5,
		DefaultAbsorptivity:  0.5,
		ACResistanceMethod:   ACResistance_IRAM2212,
		ConvectionMethod:     Convection_IEEE738,
	},
	"ASTM_B232": {
		Name:                 "ASTM B232 / IEEE 738",
		RatingWindAngleDeg:   90.0,
		RatingSolarElevation: 30.0,
		RatingSolarRadiation: 960.0,
		DefaultEmissivity:    0.5,
		DefaultAbsorptivity:  0.5,
		ACResistanceMethod:   ACResistance_ASTMB232,
		ConvectionMethod:     Convection_IEEE738,
	},
	"IEEE_738": {
		Name:                 "IEEE 738 Generic",
		RatingWindAngleDeg:   90.0,
		RatingSolarElevation: 90.0,
		RatingSolarRadiation: 1000.0,
		DefaultEmissivity:    0.9,
		DefaultAbsorptivity:  0.9,
		ACResistanceMethod:   ACResistance_IEEE738,
		ConvectionMethod:     Convection_IEEE738,
	},
}

func GetStandard(name string) (StandardConfig, bool) {
	std, ok := Standards[name]
	return std, ok
}

func acResistanceByStandard(method ACResistanceMethod, r20, alpha, maxTemp, tempAir, diameter, steelCoreDia, alAreaMM2, steelAreaMM2, pressurePa float64) float64 {
	switch method {
	case ACResistance_IRAM2187:
		return acResistanceIRAM2187(r20, alpha, maxTemp, tempAir, diameter, steelCoreDia, alAreaMM2, steelAreaMM2, pressurePa)
	case ACResistance_IRAM2212:
		return acResistanceIRAM2212(r20, alpha, maxTemp, tempAir, diameter, steelCoreDia, alAreaMM2, steelAreaMM2, pressurePa)
	case ACResistance_ASTMB232:
		return acResistanceASTMB232(r20, alpha, maxTemp, tempAir, diameter, steelCoreDia, alAreaMM2, steelAreaMM2, pressurePa)
	default:
		return acResistanceIEEE738(r20, alpha, maxTemp, tempAir, diameter, steelCoreDia, alAreaMM2, steelAreaMM2, pressurePa)
	}
}

func acResistanceIRAM2187(r20, alpha, maxTemp, tempAir, diameter, steelCoreDia, alAreaMM2, steelAreaMM2, pressurePa float64) float64 {
	rDC := r20 * (1 + alpha*(maxTemp-20.0))
	if steelCoreDia > 0 && alAreaMM2 > 0 && steelAreaMM2 > 0 {
		skinFactor := skinEffectFactorAl(maxTemp, diameter, steelCoreDia, alAreaMM2)
		proximityFactor := proximityEffectFactor(maxTemp, diameter, alAreaMM2, steelAreaMM2)
		return rDC * skinFactor * proximityFactor
	}
	return rDC * 1.02
}

func acResistanceIRAM2212(r20, alpha, maxTemp, tempAir, diameter, steelCoreDia, alAreaMM2, steelAreaMM2, pressurePa float64) float64 {
	rDC := r20 * (1 + alpha*(maxTemp-20.0))
	return rDC * 1.02
}

func acResistanceASTMB232(r20, alpha, maxTemp, tempAir, diameter, steelCoreDia, alAreaMM2, steelAreaMM2, pressurePa float64) float64 {
	return acResistanceIEEE738(r20, alpha, maxTemp, tempAir, diameter, steelCoreDia, alAreaMM2, steelAreaMM2, pressurePa)
}

func convectionByStandard(method ConvectionMethod, windAngleRad float64, reynolds, pr float64) float64 {
	switch method {
	case Convection_IRAM2187:
		return nusseltForcedFlowIEEE738(reynolds, pr)
	default:
		kAngle := kAngleFactor(windAngleRad)
		return kAngle * nusseltForcedFlowIEEE738(reynolds, pr)
	}
}
