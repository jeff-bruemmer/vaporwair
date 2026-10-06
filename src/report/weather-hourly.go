package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

func WeatherHourly(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Hourly Summary"))
	if w.Hourly.Summary != "" {
		fmt.Println(AddPeriod(w.Hourly.Summary))
		fmt.Println()
	}

	// Units live in the second header row so cells hold bare numbers and the table fits 80 columns.
	fmt.Fprintf(Table, "Time\tTemp\tFeels\tDew\tPrecip\tHumid\tWind\tGust\n")
	fmt.Fprintf(Table, "\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		temperatureUnit, temperatureUnit, temperatureUnit, percentUnit, percentUnit, windSpeedUnit, windSpeedUnit)
	for _, h := range LimitData(upcomingHours(w.Hourly.Data), DefaultHourlyLimit) {
		gustStr := "-"
		if h.WindGust > 0 {
			gustStr = fmt.Sprintf("%.0f", h.WindGust)
		}
		periodLabel := h.PeriodName
		if periodLabel == "" {
			periodLabel = FormatTime(h.Time)
		}

		fmt.Fprintf(Table, "%s\t%.0f\t%.0f\t%.0f\t%.0f\t%.0f\t%.0f\t%s\n",
			periodLabel,
			h.Temperature,
			h.ApparentTemperature,
			h.DewPoint,
			ToPercent(h.PrecipProbability),
			ToPercent(h.Humidity),
			h.WindSpeed,
			gustStr)
	}
	Table.Flush()
}
