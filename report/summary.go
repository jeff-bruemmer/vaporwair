package report

import (
	"github.com/jeff-bruemmer/vaporwair/air"
	"github.com/jeff-bruemmer/vaporwair/weather"
)

// The default report.
func Summary(w weather.Forecast, a []air.Forecast) {
	DailySummary(w)
	CurrentTemp(w)
	MinTemp(w)
	MaxTemp(w)
	Humidity(w)
	Windspeed(w)
	AirQualityIndex(a)
	Precipitation(w)
	Sunrise(w)
	Sunset(w)
}
