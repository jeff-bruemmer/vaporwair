package report

import (
	"github.com/jeff-bruemmer/vaporwair/weather"
	"github.com/jeff-bruemmer/vaporwair/air"
)

// The default report.
func Summary(w weather.Forecast, a []air.Forecast) {
	CurrentTemp(w)
	MinTemp(w)
	MaxTemp(w)
	AirQualityIndex(a)
	Precipitation(w)
	Sunrise(w)
	Sunset(w)
}
