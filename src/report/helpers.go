package report

import (
	"strings"
	"time"

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
