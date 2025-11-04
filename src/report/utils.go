package report

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// GetCurrentHourlyData finds the most current hourly data point (closest to now).
// If no future hourly data is available, it returns the Currently data point.
func GetCurrentHourlyData(w weather.Forecast) weather.DataPoint {
	if len(w.Hourly.Data) == 0 {
		return w.Currently
	}

	currentTime := float64(time.Now().Unix())
	current := w.Currently
	minDiff := float64(999999999)

	for _, hour := range w.Hourly.Data {
		diff := hour.Time - currentTime
		if diff >= 0 && diff < minDiff {
			current = hour
			minDiff = diff
		}
	}

	return current
}

// CapitalizeFirst capitalizes the first letter of a string.
// Handles special cases like "rain/snow" -> "Rain/Snow".
func CapitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	// Handle special case for mixed precipitation
	if s == "rain/snow" {
		return "Rain/Snow"
	}

	runes := []rune(s)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] = runes[0] - 32
	}

	return string(runes)
}

// GetPrecipTypeOrDefault returns a capitalized precipitation type string,
// or "Precipitation" as the default if the input is empty.
func GetPrecipTypeOrDefault(precipType string) string {
	if precipType != "" {
		return CapitalizeFirst(precipType)
	}
	return "Precipitation"
}

// SafeSliceHourly safely slices hourly data with bounds checking.
// Returns a slice of at most maxHours elements, or fewer if the data
// doesn't contain that many elements. Returns an empty slice if maxHours <= 0.
func SafeSliceHourly(data []weather.DataPoint, maxHours int) []weather.DataPoint {
	if maxHours <= 0 {
		return []weather.DataPoint{}
	}

	if len(data) < maxHours {
		maxHours = len(data)
	}

	if maxHours == 0 {
		return []weather.DataPoint{}
	}

	return data[:maxHours]
}

// IsPrecipitationNote checks if a note string contains precipitation-related keywords.
// This is used to filter out precipitation-related notes when they're displayed elsewhere.
func IsPrecipitationNote(note string) bool {
	lowerNote := strings.ToLower(note)
	precipKeywords := []string{"precipitation", "rain", "snow", "umbrella", "sleet"}

	for _, keyword := range precipKeywords {
		if strings.Contains(lowerNote, keyword) {
			return true
		}
	}

	return false
}

// FormatTemperatureWithFeelsLike prints temperature with "feels like" if the difference exceeds threshold.
// Uses FeelsLikeDiffThreshold (3°F) to determine when to show the "feels like" temperature.
func FormatTemperatureWithFeelsLike(tw *tabwriter.Writer, label string, actual, feelsLike float64, unit string) {
	diff := feelsLike - actual
	if diff > FeelsLikeDiffThreshold || diff < -FeelsLikeDiffThreshold {
		fmt.Fprintf(tw, "%s:\t%.0f %s (feels like %.0f %s)\n", label, actual, unit, feelsLike, unit)
	} else {
		fmt.Fprintf(tw, "%s:\t%.0f %s\n", label, actual, unit)
	}
}

// GetHighestAQIForToday finds the maximum AQI value from today's air quality forecasts.
// Returns the AQI value, pollutant particle name, and category description.
// Returns (-1, "", "") if no valid data is available.
func GetHighestAQIForToday(a []air.Forecast) (aqi int, particle, category string) {
	if len(a) == 0 {
		return -1, "", ""
	}

	today := a[0].DateForecast
	aqi = -1

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

	return aqi, particle, category
}
