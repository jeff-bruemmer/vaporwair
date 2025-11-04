package report

// Temperature thresholds for "feels like" display
const (
	// FeelsLikeDiffThreshold is the temperature difference (in °F) at which
	// we display "feels like" temperature separately from actual temperature
	FeelsLikeDiffThreshold = 3.0
)

// Outfit temperature ranges (in °F)
const (
	OutfitExtremeColdMax = 14 // Below this: extreme cold gear
	OutfitHeavyWinterMax = 24 // 15-24: heavy winter coat
	OutfitWinterMax      = 34 // 25-34: winter coat
	OutfitHeavyJacketMax = 44 // 35-44: heavy jacket
	OutfitWarmLayersMax  = 54 // 45-54: warm layers
	OutfitModerateMax    = 64 // 55-64: moderate layers
	OutfitLightLayersMax = 74 // 65-74: light layers
	OutfitSummerMax      = 84 // 75-84: summer wear
	// 85+: light, breathable clothing
)

// Precipitation thresholds
const (
	// PrecipSignificantThreshold is the probability at which precipitation
	// is considered significant enough to highlight (30%)
	PrecipSignificantThreshold = 0.3
)

// Default array limits for data display
const (
	DefaultHourlyLimit = 12 // Default number of hours to display
	DefaultMaxHours    = 6  // Default maximum hours for detailed view
)

// Outfit descriptions for each temperature range
const (
	OutfitDescExtremeCold = "Extreme cold gear, heavy insulation, thermal base layers"
	OutfitDescHeavyWinter = "Heavy winter coat, multiple layers, thermal underwear"
	OutfitDescWinter      = "Winter coat, insulated layers, thermal wear"
	OutfitDescHeavyJacket = "Heavy jacket or coat with layers underneath"
	OutfitDescWarmLayers  = "Warm layers (jacket, long sleeves, jeans)"
	OutfitDescModerate    = "Moderate layers (pants, sweater or light jacket)"
	OutfitDescLightLayers = "Light layers (jeans, long sleeves or light sweater)"
	OutfitDescSummer      = "Summer wear (shorts or light pants, short sleeves)"
	OutfitDescLightBreath = "Light, breathable clothing (shorts, t-shirt, tank top)"
)

// OutfitLevel represents a temperature range and its corresponding outfit recommendation.
type OutfitLevel struct {
	MinTemp int
	MaxTemp int
	Outfit  string
}

// OutfitLevels defines all temperature ranges and their corresponding outfit recommendations.
// Ordered from hottest to coldest for easy iteration.
var OutfitLevels = []OutfitLevel{
	{OutfitSummerMax + 1, 999, OutfitDescLightBreath},
	{OutfitLightLayersMax + 1, OutfitSummerMax, OutfitDescSummer},
	{OutfitModerateMax + 1, OutfitLightLayersMax, OutfitDescLightLayers},
	{OutfitWarmLayersMax + 1, OutfitModerateMax, OutfitDescModerate},
	{OutfitHeavyJacketMax + 1, OutfitWarmLayersMax, OutfitDescWarmLayers},
	{OutfitWinterMax + 1, OutfitHeavyJacketMax, OutfitDescHeavyJacket},
	{OutfitHeavyWinterMax + 1, OutfitWinterMax, OutfitDescWinter},
	{OutfitExtremeColdMax + 1, OutfitHeavyWinterMax, OutfitDescHeavyWinter},
	{-100, OutfitExtremeColdMax, OutfitDescExtremeCold},
}
