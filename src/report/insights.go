package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"text/tabwriter"
)

// generateInsightsContent generates the insights report content to a given tabwriter
func generateInsightsContent(tw *tabwriter.Writer, w weather.Forecast, a []air.Forecast, includeHeadings bool) {
	if includeHeadings {
		fmt.Fprintln(tw, Title("Comparative Analysis & Insights"))
	}
	fmt.Fprintln(tw, "")

	// Temperature comparison
	if len(w.Daily.Data) > 1 {
		today := w.Daily.Data[0]
		tomorrow := w.Daily.Data[1]

		tempDiff := tomorrow.TemperatureMax - today.TemperatureMax
		fmt.Fprintf(tw, "Tomorrow vs Today:\t%.0f%s (", tempDiff, temperatureUnit)
		if tempDiff > 0 {
			fmt.Fprint(tw, "warmer)")
		} else if tempDiff < 0 {
			fmt.Fprint(tw, "cooler)")
		} else {
			fmt.Fprint(tw, "same)")
		}
		fmt.Fprintln(tw, "")
	}

	// Current conditions - need to pass tw to these functions
	fmt.Fprintf(tw, formatValueWithUnit, "Current Temperature", Round(w.Hourly.Data[0].Temperature), temperatureUnit)
	fmt.Fprintf(tw, formatValueWithUnit, "Min Temperature", Round(w.Daily.Data[0].TemperatureMin), temperatureUnit)
	fmt.Fprintf(tw, formatValueWithUnit, "Max Temperature", w.Daily.Data[0].TemperatureMax, temperatureUnit)
	if w.Daily.Data[0].TemperatureTrend != "" {
		fmt.Fprintf(tw, formatLabelValue, "Temp Trend", w.Daily.Data[0].TemperatureTrend, "")
	}

	// Hourly trend analysis
	if len(w.Hourly.Data) >= 3 {
		fmt.Fprintln(tw, "")
		if includeHeadings {
			fmt.Fprintln(tw, Title("Next Few Hours"))
		}
		for i := 0; i < 3 && i < len(w.Hourly.Data); i++ {
			hour := w.Hourly.Data[i]
			fmt.Fprintf(tw, "%s:\t%.0f%s - %s\n",
				FormatTime(hour.Time),
				Round(hour.Temperature),
				temperatureUnit,
				hour.Summary)
		}
	}

	// Week overview
	fmt.Fprintln(tw, "")
	if includeHeadings {
		fmt.Fprintln(tw, Title("Week Overview"))
	}
	fmt.Fprintf(tw, formatString, "This week", AddPeriod(w.Daily.Summary))

	// Additional metrics
	fmt.Fprintf(tw, formatValueWithUnit, "Precipitation", Round(ToPercent(w.Daily.Data[0].PrecipProbability)), percentUnit)

	windDir := DegreesToCardinal(w.Currently.WindBearing)
	fmt.Fprintf(tw, "Windspeed:\t%.0f %s from %s\n", w.Currently.WindSpeed, windSpeedUnit, windDir)
	if w.Currently.WindGust > 0 {
		fmt.Fprintf(tw, formatValueWithUnit, "Wind Gust", w.Currently.WindGust, windSpeedUnit)
	}

	fmt.Fprintf(tw, formatValueWithUnit, "Pressure", w.Daily.Data[0].Pressure, pressureUnit)
	fmt.Fprintf(tw, formatValueWithUnit, "Visibility", w.Daily.Data[0].Visibility, distanceUnit)
	if w.Daily.Data[0].DewPoint > 0 {
		fmt.Fprintf(tw, formatValueWithUnit, "Dew Point", w.Daily.Data[0].DewPoint, temperatureUnit)
	}

	// Air Quality Index
	if len(a) == 0 {
		fmt.Fprintf(tw, formatMultipleValues, "Air Quality Index", "N/A", "No data", "unavailable")
	} else {
		today := a[0].DateForecast
		aqi := -1
		var particle string
		var category string
		var categoryOnly string

		for _, measurement := range a {
			if measurement.DateForecast != today {
				break
			}
			if measurement.Category.Name != "" && categoryOnly == "" {
				categoryOnly = measurement.Category.Name
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
			fmt.Fprintf(tw, formatMultipleValues, "Air Quality Index", aqi, particle, category)
		} else if categoryOnly != "" {
			fmt.Fprintf(tw, formatMultipleValues, "Air Quality", categoryOnly, "(numeric", "forecast pending)")
		} else {
			fmt.Fprintf(tw, formatMultipleValues, "Air Quality Index", "N/A", "Forecast", "not yet available")
		}
	}

	fmt.Fprintf(tw, formatNumber, "UV Index", w.Currently.UVIndex)
}

// InsightsReport provides comparative analysis and time-based insights.
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

	// Clothing recommendations
	fmt.Fprintln(TW, "")
	ClothingSummary(w, a)
	fmt.Println()
}
