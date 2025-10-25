package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// The default report.
func Summary(w weather.Forecast, a []air.Forecast) {
	WeeklySummary(w)
	DailySummary(w)
	CurrentTemp(w)
	MinTemp(w)
	MaxTemp(w)
	Humidity(w)
	Windspeed(w)
	AirQualityIndex(a)
	UVIndex(w)
	Precipitation(w)
	TW.Flush()

	// Add clothing recommendations
	fmt.Println()
	ClothingSummary(w, a)
}
