package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// InsightsReport provides comparative analysis and time-based insights.
func InsightsReport(w weather.Forecast, a []air.Forecast) {
	// Show weather alerts first if any exist
	WeatherAlerts(w)

	fmt.Fprintln(TW, Title("Comparative Analysis & Insights"))
	fmt.Fprintln(TW, "")

	// Temperature comparison
	if len(w.Daily.Data) > 1 {
		today := w.Daily.Data[0]
		tomorrow := w.Daily.Data[1]

		tempDiff := tomorrow.TemperatureMax - today.TemperatureMax
		fmt.Fprintf(TW, "Tomorrow vs Today:\t%.0f%s (", tempDiff, temperatureUnit)
		if tempDiff > 0 {
			fmt.Fprint(TW, "warmer)")
		} else if tempDiff < 0 {
			fmt.Fprint(TW, "cooler)")
		} else {
			fmt.Fprint(TW, "same)")
		}
		fmt.Fprintln(TW, "")
	}

	// Current conditions
	CurrentTemp(w)
	MinTemp(w)
	MaxTemp(w)

	// Hourly trend analysis
	if len(w.Hourly.Data) >= 3 {
		fmt.Fprintln(TW, "")
		fmt.Fprintln(TW, Title("Next Few Hours"))
		for i := 0; i < 3 && i < len(w.Hourly.Data); i++ {
			hour := w.Hourly.Data[i]
			fmt.Fprintf(TW, "%s:\t%.0f%s - %s\n",
				FormatTime(hour.Time),
				Round(hour.Temperature),
				temperatureUnit,
				hour.Summary)
		}
	}

	// Week overview
	fmt.Fprintln(TW, "")
	fmt.Fprintln(TW, Title("Week Overview"))
	WeeklySummary(w)

	// Additional metrics
	Precipitation(w)
	Windspeed(w)
	Pressure(w)
	Visibility(w)
	AirQualityIndex(a)
	UVIndex(w)

	TW.Flush()

	// Clothing recommendations
	fmt.Fprintln(TW, "")
	ClothingSummary(w, a)
}
