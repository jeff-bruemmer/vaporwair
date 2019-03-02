package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/air"
	"github.com/jeff-bruemmer/vaporwair/weather"
)

func WeatherHourly(w weather.Forecast, a []air.Forecast) {
	//format := "%v\t%.0f %s\t%.0f %s\t%.0f %s\t%.0f %s\t%.0f %s\n"
	format := "%v\t%.0f %s\t%.0f %s\t%.0f %s\t%.2f %s\t%.0f %s\n"
	fmt.Fprintf(TW, "Hour\tTemp\tFeels Like\tPrecip\tIntensity\tWind\n")
	fmt.Fprintf(TW, "----\t----\t----------\t------\t---------\t----\n")
	d := LimitData(w.Hourly.Data, 12)
	for _, h := range d {
		fmt.Fprintf(TW, format,
			FormatTime(h.Time),
			h.Temperature, tu,
			h.ApparentTemperature, tu,
			ToPercent(h.PrecipProbability), pc,
			ToPercent(h.PrecipIntensity), "mmph",
			h.WindSpeed, wu)
	}
	TW.Flush()
}
