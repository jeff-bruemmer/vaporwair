package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

func WeatherHourly(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Hourly Summary"))
	fmt.Println(AddPeriod(w.Hourly.Summary))
	fmt.Println()
	format := "%v\t%.0f %s\t%.0f %s\t%.0f %s\t%s\n"
	fmt.Fprintf(TW, "Hour\tTemp\tPrecip\tWind\tGust\n")
	fmt.Fprintf(TW, "----\t----\t------\t----\t----\n")
	d := LimitData(w.Hourly.Data, 12)
	for _, h := range d {
		gustStr := "-"
		if h.WindGust > 0 {
			gustStr = fmt.Sprintf("%.0f %s", h.WindGust, windSpeedUnit)
		}
		fmt.Fprintf(TW, format,
			FormatTime(h.Time),
			h.Temperature, temperatureUnit,
			ToPercent(h.PrecipProbability), percentUnit,
			h.WindSpeed, windSpeedUnit,
			gustStr)
	}
	TW.Flush()
}
