package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// WeatherDaily displays comprehensive daily forecast with all available NOAA fields.
func WeatherDaily(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Daily Forecast"))

	// Show overall summary if available
	if w.Daily.Summary != "" {
		fmt.Println(AddPeriod(w.Daily.Summary))
		fmt.Println()
	}

	data := LimitData(w.Daily.Data, 7)

	for i, day := range data {
		if i > 0 {
			fmt.Println()
		}

		// Day header
		dayLabel := day.PeriodName
		if dayLabel == "" {
			dayLabel = time.Unix(int64(day.Time), 0).Format("Monday, Jan 2")
		}
		fmt.Println(Bold(dayLabel))

		// Temperature information (0F and below are real readings, so no "> 0" guards)
		fmt.Fprintf(TW, formatValueWithUnit, "High", day.TemperatureMax, temperatureUnit)
		fmt.Fprintf(TW, formatValueWithUnit, "Low", day.TemperatureMin, temperatureUnit)
		if day.TemperatureTrend != "" {
			fmt.Fprintf(TW, formatString, "Temperature Trend", day.TemperatureTrend)
		}

		// Precipitation
		if day.PrecipProbability > 0 {
			precipType := day.PrecipType
			if precipType != "" {
				fmt.Fprintf(TW, "%s Chance:\t%.0f%s\n",
					CapitalizeFirst(precipType),
					ToPercent(day.PrecipProbability),
					percentUnit)
			} else {
				fmt.Fprintf(TW, formatValueWithUnit, "Precipitation Chance",
					ToPercent(day.PrecipProbability), percentUnit)
			}
		}

		// Wind information
		if day.WindSpeed > 0 {
			windStr := FormatWindString(day.WindSpeed, day.WindBearing, day.WindGust, windSpeedUnit, true)
			fmt.Fprintf(TW, "Wind:\t%s\n", windStr)
		}

		// Humidity and dewpoint
		if day.Humidity > 0 {
			fmt.Fprintf(TW, formatValueWithUnit, "Humidity",
				ToPercent(day.Humidity), percentUnit)
		}
		// NOAA's daily periods usually omit dewpoint (stored as 0). A converted Celsius
		// reading is essentially never exactly 0F, so 0 here means missing.
		if day.DewPoint != 0 {
			fmt.Fprintf(TW, formatValueWithUnit, "Dewpoint", day.DewPoint, temperatureUnit)
		}

		// Pressure and visibility come from a live observation, so only today has them
		if day.Pressure > 0 {
			fmt.Fprintf(TW, formatPressure, "Pressure", day.Pressure, pressureUnit)
		}
		if day.Visibility > 0 {
			fmt.Fprintf(TW, formatValueWithWordUnit, "Visibility", day.Visibility, distanceUnit)
		}

		// Detailed forecast, in the same flush so it aligns with the fields above
		if day.DetailedForecast != "" {
			maxWidth := calculateValueColumnWidth("Forecast")
			wrappedForecast := strings.Join(WrapText(AddPeriod(day.DetailedForecast), maxWidth), "\n\t")
			fmt.Fprintf(TW, formatString, "Forecast", wrappedForecast)
		} else if day.Summary != "" {
			maxWidth := calculateValueColumnWidth("Summary")
			wrappedSummary := strings.Join(WrapText(AddPeriod(day.Summary), maxWidth), "\n\t")
			fmt.Fprintf(TW, formatString, "Summary", wrappedSummary)
		}
		TW.Flush()
	}
}
