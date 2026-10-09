package report

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// GetCurrentHourlyData finds the hour in progress, or failing that the nearest future hour.
// If no current or future hourly data is available, it returns the Currently data point.
func GetCurrentHourlyData(w weather.Forecast) weather.DataPoint {
	if len(w.Hourly.Data) == 0 {
		return w.Currently
	}

	currentTime := float64(clock().Unix())
	for _, hour := range w.Hourly.Data {
		if hour.Time <= currentTime && currentTime < hour.Time+3600 {
			return hour
		}
	}
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

// FormatTemperatureString returns a temperature string with "feels like" if the difference exceeds threshold.
// Uses FeelsLikeDiffThreshold (3°F) to determine when to show the "feels like" temperature.
func FormatTemperatureString(actual, feelsLike float64, unit string) string {
	diff := feelsLike - actual
	if diff > FeelsLikeDiffThreshold || diff < -FeelsLikeDiffThreshold {
		return fmt.Sprintf("%.0f%s (feels %.0f%s)", actual, unit, feelsLike, unit)
	}
	return fmt.Sprintf("%.0f%s", actual, unit)
}

// FormatTemperatureWithFeelsLike prints temperature with "feels like" if the difference exceeds threshold.
// Uses FeelsLikeDiffThreshold (3°F) to determine when to show the "feels like" temperature.
func FormatTemperatureWithFeelsLike(tw *tabwriter.Writer, label string, actual, feelsLike float64, unit string) {
	diff := feelsLike - actual
	if diff > FeelsLikeDiffThreshold || diff < -FeelsLikeDiffThreshold {
		fmt.Fprintf(tw, "%s:\t%.0f%s (feels like %.0f%s)\n", label, actual, unit, feelsLike, unit)
	} else {
		fmt.Fprintf(tw, "%s:\t%.0f%s\n", label, actual, unit)
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

// WrapText wraps text to specified width, breaking on word boundaries.
// Width is measured in runes so multi-byte characters like ° count as one column.
// Returns a slice of strings, one for each line.
func WrapText(text string, width int) []string {
	if utf8.RuneCountInString(text) <= width {
		return []string{text}
	}

	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if utf8.RuneCountInString(currentLine)+1+utf8.RuneCountInString(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// FormatWindString formats wind speed and direction with optional gust information.
// If withFrom is true, uses "from {direction}" format, otherwise uses "{direction}" format.
func FormatWindString(speed, bearing, gust float64, unit string, withFrom bool) string {
	dir := DegreesToCardinal(bearing)
	if withFrom {
		dir = "from " + dir
	}
	s := fmt.Sprintf("%.0f %s %s", speed, unit, dir)
	if gust > 0 {
		s += fmt.Sprintf(" (gusts %.0f)", gust)
	}
	return s
}

// DegreesToCardinal converts wind bearing in degrees to cardinal direction.
func DegreesToCardinal(degrees float64) string {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	index := int((degrees + 11.25) / 22.5)
	return directions[index%16]
}

// LabeledText is one labeled section of a longer piece of text.
type LabeledText struct {
	Label string
	Text  string
}

// ParseNWSDescription splits an NWS alert description of the form
// "* WHAT...text * WHERE...text" into labeled sections ("What", "Where").
// Returns nil if the description doesn't follow that structure.
func ParseNWSDescription(desc string) []LabeledText {
	normalized := strings.Join(strings.Fields(desc), " ")
	if !strings.HasPrefix(normalized, "* ") {
		return nil
	}

	var sections []LabeledText
	for _, part := range strings.Split(normalized[2:], " * ") {
		label, text, ok := strings.Cut(part, "...")
		if !ok || label == "" || label != strings.ToUpper(label) {
			return nil
		}
		sections = append(sections, LabeledText{
			Label: CapitalizeFirst(strings.ToLower(label)),
			Text:  strings.TrimSpace(text),
		})
	}
	return sections
}

// FormatUntil formats a future time as clock time with a relative offset, e.g. "Wed 05:00 (in 15h)".
// The weekday is omitted when t falls on the same day as now.
func FormatUntil(t, now time.Time) string {
	at := clockAt(t, now)
	d := t.Sub(now)
	if d <= 0 {
		return at
	}
	return fmt.Sprintf("%s (in %s)", at, FormatDuration(d))
}

// clockAt formats t as "15:04", or "Mon 15:04" when it falls on a different day than ref.
func clockAt(t, ref time.Time) string {
	if t.YearDay() != ref.YearDay() || t.Year() != ref.Year() {
		return t.Format("Mon 15:04")
	}
	return t.Format("15:04")
}

// AlertHeadline marks an alert title in plain ASCII, e.g. "! WINTER STORM WARNING".
// Output is black and white, so the marker and capitals carry the emphasis.
func AlertHeadline(title string) string {
	return "! " + strings.ToUpper(title)
}

// alertHeadlineWithWindow is an alert's headline followed by when it applies, e.g.
// "! FROST ADVISORY | Wed 00:00-05:00 (starts in 4h)".
func alertHeadlineWithWindow(a weather.Alert, now time.Time) string {
	headline := AlertHeadline(a.Title)
	if window := FormatAlertWindow(unixOrZero(a.Time), unixOrZero(a.Expires), now); window != "" {
		headline += " | " + window
	}
	return headline
}

// aqiConcern lists the EPA categories at which outdoor activity should be limited.
var aqiConcern = map[string]bool{
	"Unhealthy for Sensitive Groups": true,
	"Unhealthy":                      true,
	"Very Unhealthy":                 true,
	"Hazardous":                      true,
}

// AQICategory returns the EPA category name, with a trailing " !" when it is
// "Unhealthy for Sensitive Groups" or worse, so severity reads without color.
func AQICategory(name string) string {
	if aqiConcern[name] {
		return name + " !"
	}
	return name
}

// clock is the report's notion of now; tests replace it to render a fixed time of day.
var clock = time.Now

// clothingWindowEnd is when the hours you'd dress for end: midnight, so an evening
// recommendation isn't driven by tomorrow's pre-dawn low. Late at night the window
// still spans at least ClothingMinHours, and never more than ClothingHours.
func clothingWindowEnd(now time.Time) time.Time {
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	end := min(midnight.Unix(), now.Add(ClothingHours*time.Hour).Unix())
	end = max(end, now.Add(ClothingMinHours*time.Hour).Unix())
	return time.Unix(end, 0).In(now.Location())
}

// windowHours returns the hourly data points from the hour in progress up to end.
func windowHours(w weather.Forecast, end time.Time) []weather.DataPoint {
	var hours []weather.DataPoint
	for _, h := range upcomingHours(w.Hourly.Data) {
		if h.Time >= float64(end.Unix()) {
			break
		}
		hours = append(hours, h)
	}
	return hours
}

// FeelsLikeRange returns the coldest feels-like temperature (and the hour it occurs)
// and the warmest actual temperature over hours. It falls back to the first daily
// period's low and high, with a zero coldestAt, when hours is empty.
func FeelsLikeRange(w weather.Forecast, hours []weather.DataPoint) (coldest, warmest, coldestAt float64) {
	if len(hours) == 0 {
		if len(w.Daily.Data) > 0 {
			return w.Daily.Data[0].TemperatureMin, w.Daily.Data[0].TemperatureMax, 0
		}
		return 0, 0, 0
	}
	coldest, warmest, coldestAt = hours[0].ApparentTemperature, hours[0].Temperature, hours[0].Time
	for _, h := range hours[1:] {
		if h.ApparentTemperature < coldest {
			coldest, coldestAt = h.ApparentTemperature, h.Time
		}
		warmest = max(warmest, h.Temperature)
	}
	return coldest, warmest, coldestAt
}

// IsNightPeriod reports whether a daily period is overnight. After about 6pm NOAA's first
// period is one, and it has no daytime high. Forecasts cached before Night was recorded
// fall back to the period name ("Tonight", "Overnight", "Monday Night").
func IsNightPeriod(day weather.DataPoint) bool {
	name := day.PeriodName
	return day.Night || name == "Tonight" || name == "Overnight" || strings.HasSuffix(name, " Night")
}

// FormatAlertWindow describes when an alert applies relative to now:
// "until Wed 05:00 (in 9h)" once it is in effect, or "Wed 00:00-05:00 (starts in 4h)"
// before its onset. A zero onset or expiry is left out.
func FormatAlertWindow(onset, expires, now time.Time) string {
	if onset.IsZero() || !onset.After(now) {
		if expires.IsZero() {
			return ""
		}
		return "until " + FormatUntil(expires, now)
	}
	window := clockAt(onset, now)
	if !expires.IsZero() {
		window += "-" + clockAt(expires, onset)
	}
	return fmt.Sprintf("%s (starts in %s)", window, FormatDuration(onset.Sub(now)))
}

// unixOrZero converts a Unix timestamp to a time, mapping 0 (missing) to the zero time.
func unixOrZero(t float64) time.Time {
	if t <= 0 {
		return time.Time{}
	}
	return time.Unix(int64(t), 0)
}

// FormatDuration renders a positive duration compactly: "45m", "4h", "2d".
func FormatDuration(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// Truncate shortens s to at most n runes, ending in "..." when cut.
func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 3 {
		return string(r[:n])
	}
	return string(r[:n-3]) + "..."
}

// upcomingHours drops hours that have already ended; NOAA's hourly feed can lag by an hour or two.
func upcomingHours(data []weather.DataPoint) []weather.DataPoint {
	hourStart := float64(clock().Add(-time.Hour).Unix())
	for i, h := range data {
		if h.Time > hourStart {
			return data[i:]
		}
	}
	return nil
}

// DropPast removes what a forecast has outlived by now: expired alerts, daily periods that
// have ended, and air forecasts for earlier days. A cache shown offline can be a day old.
// It never writes to the slices it is given, since the caller saves them afterwards.
func DropPast(w weather.Forecast, a []air.Forecast, now time.Time) (weather.Forecast, []air.Forecast) {
	// A period has ended once the next one has started. The last is kept, so Data[0] exists.
	for len(w.Daily.Data) > 1 && w.Daily.Data[1].Time <= float64(now.Unix()) {
		w.Daily.Data = w.Daily.Data[1:]
	}

	var alerts []weather.Alert
	for _, alert := range w.Alerts {
		if alert.Expires == 0 || alert.Expires > float64(now.Unix()) {
			alerts = append(alerts, alert)
		}
	}
	w.Alerts = alerts

	// AirNow dates are the location's, so "today" is too. LoadLocation("") would mean UTC.
	if w.Timezone != "" {
		if loc, err := time.LoadLocation(w.Timezone); err == nil {
			now = now.In(loc)
		}
	}
	today := now.Format("2006-01-02")
	var current []air.Forecast
	for _, f := range a {
		if strings.TrimSpace(f.DateForecast) >= today {
			current = append(current, f)
		}
	}
	return w, current
}
