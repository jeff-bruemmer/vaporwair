package report

import (
	"fmt"
	"strings"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"text/tabwriter"
	"time"
)

// generateInsightsContent generates the insights report content to a given tabwriter
func generateInsightsContent(tw *tabwriter.Writer, w weather.Forecast, a []air.Forecast, includeHeadings bool) {
	daily := w.Daily.Data[0]

	// Find the most current hourly data point (closest to now)
	current := GetCurrentHourlyData(w)

	// Today's Summary
	if includeHeadings {
		fmt.Fprintln(tw, Title("Today's Forecast"))
	}

	// Show detailed forecast for today if available
	if daily.DetailedForecast != "" {
		wrapped := wrapTextSimple(AddPeriod(daily.DetailedForecast), HeadingWidth-4)
		fmt.Fprintf(tw, "%s\n", wrapped)
	} else if daily.Summary != "" {
		wrapped := wrapTextSimple(AddPeriod(daily.Summary), HeadingWidth-4)
		fmt.Fprintf(tw, "%s\n", wrapped)
	}
	fmt.Fprintln(tw, "")

	// Current conditions
	if includeHeadings {
		fmt.Fprintln(tw, Title("Current Conditions"))
	}

	// Temperature with feels like
	actual := Round(current.Temperature)
	feelsLike := Round(current.ApparentTemperature)
	diff := feelsLike - actual
	if diff > FeelsLikeDiffThreshold || diff < -FeelsLikeDiffThreshold {
		fmt.Fprintf(tw, "Temperature:\t%.0f %s (feels like %.0f %s)\n",
			actual, temperatureUnit, feelsLike, temperatureUnit)
	} else {
		fmt.Fprintf(tw, formatValueWithUnit, "Temperature", actual, temperatureUnit)
	}

	fmt.Fprintf(tw, formatValueWithUnit, "Today's High", daily.TemperatureMax, temperatureUnit)
	fmt.Fprintf(tw, formatValueWithUnit, "Today's Low", Round(daily.TemperatureMin), temperatureUnit)

	// Precipitation
	if daily.PrecipProbability > 0 {
		precipType := daily.PrecipType
		if precipType != "" {
			fmt.Fprintf(tw, "%s Chance:\t%.0f %s\n",
				CapitalizeFirst(precipType),
				Round(ToPercent(daily.PrecipProbability)),
				percentUnit)
		} else {
			fmt.Fprintf(tw, formatValueWithUnit, "Precipitation", Round(ToPercent(daily.PrecipProbability)), percentUnit)
		}
	}

	// Humidity and dewpoint
	fmt.Fprintf(tw, formatValueWithUnit, "Humidity", ToPercent(current.Humidity), percentUnit)
	if current.DewPoint > 0 {
		fmt.Fprintf(tw, formatValueWithUnit, "Dewpoint", current.DewPoint, temperatureUnit)
	}

	// Wind
	windDir := DegreesToCardinal(current.WindBearing)
	fmt.Fprintf(tw, "Wind:\t%.0f %s from %s\n", current.WindSpeed, windSpeedUnit, windDir)
	if current.WindGust > 0 {
		fmt.Fprintf(tw, formatValueWithUnit, "Gusts", current.WindGust, windSpeedUnit)
	}

	// Pressure and visibility
	if current.Pressure > 0 {
		fmt.Fprintf(tw, formatValueWithUnit, "Pressure", current.Pressure, pressureUnit)
	}
	if current.Visibility > 0 {
		fmt.Fprintf(tw, formatValueWithUnit, "Visibility", current.Visibility, distanceUnit)
	}

	// Air Quality
	if len(a) > 0 {
		today := a[0].DateForecast
		aqi := -1
		var particle string
		var category string

		for _, measurement := range a {
			if measurement.DateForecast != today {
				break
			}
			if measurement.AQI < 0 {
				continue
			}
			if measurement.AQI > aqi {
				aqi = measurement.AQI
				particle = measurement.ParameterName
				category = measurement.Category.Name
			}
		}

		if aqi >= 0 {
			fmt.Fprintf(tw, "Air Quality:\t%d AQI (%s) - %s\n", aqi, particle, category)
		}
	}
}

// InsightsReport provides today's forecast, what to wear, and next few hours.
func InsightsReport(w weather.Forecast, a []air.Forecast) {
	// Show weather alerts first if any exist
	WeatherAlerts(w)

	// First pass: measure content width (without headings)
	SetHeadingWidthFromContent(func(tw *tabwriter.Writer) {
		generateInsightsContent(tw, w, a, false)
	})

	// Second pass: generate actual output with properly sized headings
	generateInsightsContent(TW, w, a, true)
	TW.Flush()

	// What to Wair
	fmt.Println()
	ClothingSummary(w, a)

	// Next few hours
	fmt.Println()
	fmt.Println(Title("Next Few Hours"))

	// Table header
	fmt.Fprintf(TW, "Time\tTemp\tConditions\tPrecip\tWind\n")
	fmt.Fprintf(TW, "----\t----\t----------\t------\t----\n")

	currentTime := float64(time.Now().Unix())
	hoursShown := 0
	maxHours := 6

	for i := 0; i < len(w.Hourly.Data) && hoursShown < maxHours; i++ {
		hour := w.Hourly.Data[i]
		// Only show hours that are in the future
		if hour.Time > currentTime {
			periodLabel := hour.PeriodName
			if periodLabel == "" {
				periodLabel = FormatTime(hour.Time)
			}

			// Format temperature with feels like if different
			tempStr := fmt.Sprintf("%.0f%s", Round(hour.Temperature), temperatureUnit)
			if hour.ApparentTemperature > 0 {
				diff := hour.ApparentTemperature - hour.Temperature
				if diff > 3 || diff < -3 {
					tempStr = fmt.Sprintf("%.0f%s (feels %.0f%s)",
						Round(hour.Temperature), temperatureUnit,
						Round(hour.ApparentTemperature), temperatureUnit)
				}
			}

			// Build precipitation string
			precipStr := ""
			if hour.PrecipProbability > 0 {
				if hour.PrecipType != "" {
					precipStr = fmt.Sprintf("%.0f%% %s", ToPercent(hour.PrecipProbability), hour.PrecipType)
				} else {
					precipStr = fmt.Sprintf("%.0f%% precip", ToPercent(hour.PrecipProbability))
				}
			} else {
				precipStr = "0%"
			}

			// Build wind string
			windDir := DegreesToCardinal(hour.WindBearing)
			windStr := fmt.Sprintf("%.0f %s %s", hour.WindSpeed, windSpeedUnit, windDir)
			if hour.WindGust > 0 {
				windStr = fmt.Sprintf("%.0f %s %s (gusts %.0f)", hour.WindSpeed, windSpeedUnit, windDir, hour.WindGust)
			}

			fmt.Fprintf(TW, "%s\t%s\t%s\t%s\t%s\n",
				periodLabel,
				tempStr,
				hour.Summary,
				precipStr,
				windStr)
			hoursShown++
		}
	}

	TW.Flush()
	fmt.Println()
}

// wrapTextSimple wraps text to specified width without adding tabs (for standalone text)
func wrapTextSimple(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	var result string
	words := strings.Fields(text)
	lineLength := 0

	for i, word := range words {
		wordLen := len(word)

		// If adding this word would exceed the width, start a new line
		if lineLength+wordLen > maxWidth && lineLength > 0 {
			result += "\n"
			lineLength = 0
		} else if i > 0 && lineLength > 0 {
			result += " "
			lineLength++
		}

		result += word
		lineLength += wordLen
	}

	return result
}
