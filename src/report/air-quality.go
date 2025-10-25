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

	// Check if we have any category data to display (even if numeric AQI is -1)
	hasData := false
	for _, f := range a {
		if f.Category.Name != "" {
			hasData = true
			break
		}
	}

	if !hasData {
		fmt.Println("\nNo air quality data available.")
		return
	}

	format := "%s\t%v\t%v\t%s\n"
	formatNoPending := "%s\t%s\t%v\t%s\n"
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

		// If we have a numeric AQI, show it
		if f.AQI >= 0 {
			fmt.Fprintf(TW, format,
				f.ParameterName,
				f.AQI,
				f.Category.Number,
				f.Category.Name)
		} else {
			// No numeric AQI yet, but show category if available
			fmt.Fprintf(TW, formatNoPending,
				f.ParameterName,
				"pending",
				f.Category.Number,
				f.Category.Name)
		}
		TW.Flush()
	}

	// Add note if any AQI values are pending
	anyPending := false
	for _, f := range a {
		if f.AQI < 0 {
			anyPending = true
			break
		}
	}
	if anyPending {
		fmt.Println("\nNote: Numeric AQI values marked 'pending' will be updated later in the day.")
	}
}
