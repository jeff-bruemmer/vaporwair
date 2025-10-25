package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// AirQuality prints AQI levels for today and tomorrow.
// Includes O3, PM2.5, PM10, NO2, and CO indices.
func AirQuality(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Air Quality Forecast"))

	// Check if any forecasts are available
	if len(a) == 0 {
		fmt.Println("\nNo air quality data available.")
		return
	}

	// Check if all AQI values are -1 (forecast not yet calculated)
	allUnavailable := true
	for _, f := range a {
		if f.AQI >= 0 {
			allUnavailable = false
			break
		}
	}

	if allUnavailable {
		fmt.Println("\nAir quality forecasts are not yet available.")
		fmt.Println("AirNow typically publishes forecasts later in the day.")
		return
	}

	format := "%s\t%v\t%v\t%s\n"
	date := ""
	fmt.Fprintf(TW, "Type\tAQI\tCategory\tDescription\n")
	fmt.Fprintf(TW, "----\t---\t--------\t-----------\n")
	for _, f := range a {
		// Skip entries with invalid AQI (-1 means not yet calculated)
		if f.AQI < 0 {
			continue
		}

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
