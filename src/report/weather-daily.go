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
		fmt.Println(Separator)
		fmt.Printf("%s\n", dayLabel)
		fmt.Println(Separator)

		// Temperature information
		if day.TemperatureMax > 0 {
			fmt.Fprintf(TW, formatValueWithUnit, "High", day.TemperatureMax, temperatureUnit)
		}
		if day.TemperatureMin > 0 {
			fmt.Fprintf(TW, formatValueWithUnit, "Low", day.TemperatureMin, temperatureUnit)
		}
		if day.TemperatureTrend != "" {
			fmt.Fprintf(TW, formatString, "Temperature Trend", day.TemperatureTrend)
		}

		// Precipitation
		if day.PrecipProbability > 0 {
			precipType := day.PrecipType
			if precipType != "" {
				fmt.Fprintf(TW, "%s Chance:\t%.0f %s\n",
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
		if day.DewPoint > 0 {
			fmt.Fprintf(TW, formatValueWithUnit, "Dewpoint", day.DewPoint, temperatureUnit)
		}

		// Pressure and visibility
		if day.Pressure > 0 {
			fmt.Fprintf(TW, formatValueWithUnit, "Pressure", day.Pressure, pressureUnit)
		}
		if day.Visibility > 0 {
			fmt.Fprintf(TW, formatValueWithUnit, "Visibility", day.Visibility, distanceUnit)
		}

		TW.Flush()

		// Detailed forecast
		if day.DetailedForecast != "" {
			maxWidth := calculateValueColumnWidth("Forecast")
			wrappedForecast := strings.Join(WrapText(AddPeriod(day.DetailedForecast), maxWidth), "\n\t")
			fmt.Fprintf(TW, formatString, "Forecast", wrappedForecast)
			TW.Flush()
		} else if day.Summary != "" {
			maxWidth := calculateValueColumnWidth("Summary")
			wrappedSummary := strings.Join(WrapText(AddPeriod(day.Summary), maxWidth), "\n\t")
			fmt.Fprintf(TW, formatString, "Summary", wrappedSummary)
			TW.Flush()
		}
	}
}
