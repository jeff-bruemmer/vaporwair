package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/air"
	"github.com/jeff-bruemmer/vaporwair/weather"
)

// AirQuality prints AQI levels for today and tomorrow.
// Includes O3, PM2.5, PM10, NO2, and CO indices.
func AirQuality(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Air Quality Forecast"))
	format := "%s\t%v\t%v\t%s\n"
	date := ""
	fmt.Fprintf(TW, "Type\tAQI\tCategory\tDescription\n")
	fmt.Fprintf(TW, "----\t---\t--------\t-----------\n")
	for _, f := range a {
		if f.DateForecast != date {
			fmt.Println()
			date = f.DateForecast
			fmt.Println(date)
			fmt.Println("==========")
		}
		fmt.Fprintf(TW, format,
			f.ParameterName,
			f.AQI,
			f.Category.Number,
			f.Category.Name)
		TW.Flush()
	}
}
