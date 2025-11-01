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
	format := "%v\t%.0f %s\t%.0f %s\t%.0f %s\t%.0f %s\t%s\n"
	fmt.Fprintf(TW, "Period\tTemp\tPrecip\tClouds\tWind\tGust\n")
	fmt.Fprintf(TW, "------\t----\t------\t------\t----\t----\n")
	d := LimitData(w.Hourly.Data, 12)
	for _, h := range d {
		gustStr := "-"
		if h.WindGust > 0 {
			gustStr = fmt.Sprintf("%.0f %s", h.WindGust, windSpeedUnit)
		}
		// Period name (e.g., "This Afternoon") is more descriptive than time alone
		periodLabel := h.PeriodName
		if periodLabel == "" {
			periodLabel = FormatTime(h.Time)
		}
		fmt.Fprintf(TW, format,
			periodLabel,
			h.Temperature, temperatureUnit,
			ToPercent(h.PrecipProbability), percentUnit,
			ToPercent(h.CloudCover), percentUnit,
			h.WindSpeed, windSpeedUnit,
			gustStr)
	}
	TW.Flush()
}
