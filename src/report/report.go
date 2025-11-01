package report

import (
	"bytes"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// Tabwriter configuration
var output = os.Stdout

const (
	minwidth = 10
	tabwidth = 0
	padding  = 2
	padchar  = ' '
	flags    = 0
)

// Unit symbols
var temperatureUnit = "°F"
var timeFormat = "HH:MM"
var windSpeedUnit = "mph"
var pressureUnit = "atm"
var distanceUnit = "miles"
var percentUnit = "%"

// Separator separates report summaries from tables.
var Separator = "+++"

var TW = tabwriter.NewWriter(output, minwidth, tabwidth, padding, padchar, flags)

// HeadingWidth is set dynamically based on content width
var HeadingWidth = 75

// Format strings for tabwriter output
var formatValueWithTime = "%s:\t%.0f %s at %v %s\n"      // e.g., "Min Temperature: 33 °F at 19:00 HH:MM"
var formatValueWithUnit = "%s:\t%.0f %s\n"               // e.g., "Humidity: 83 %"
var formatLabelValue = "%s:\t%v %s\n"                    // e.g., "Sunrise: 06:15 HH:MM"
var formatMultipleValues = "%s:\t%v %s %s\n"             // e.g., "Air Quality Index: 55 O3 Moderate"
var formatString = "%s:\t%s\n"                           // e.g., "Currently: Mostly Cloudy"
var formatNumber = "%s:\t%v\n"                           // e.g., "UV Index: 0"

// Adds title frame
func Title(t string) string {
	title := "== " + strings.ToUpper(t) + " "

	// Calculate remaining space and fill with =
	remaining := HeadingWidth - len(title)
	if remaining > 0 {
		title += strings.Repeat("=", remaining)
	} else {
		title += "=="
	}

	return title
}

// MeasureMaxLineWidth measures the maximum line width of formatted output
func MeasureMaxLineWidth(content string) int {
	maxWidth := 0
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Remove ANSI codes if any and measure visible characters
		visibleLen := len(line)
		if visibleLen > maxWidth {
			maxWidth = visibleLen
		}
	}
	return maxWidth
}

// SetHeadingWidthFromContent generates content to a buffer, measures it, and sets HeadingWidth
func SetHeadingWidthFromContent(contentGenerator func(*tabwriter.Writer)) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	tempTW := tabwriter.NewWriter(&buf, minwidth, tabwidth, padding, padchar, flags)

	// Generate content to buffer
	contentGenerator(tempTW)
	tempTW.Flush()

	// Measure max line width
	maxWidth := MeasureMaxLineWidth(buf.String())

	// Set heading width (with a minimum of 60)
	if maxWidth < 60 {
		HeadingWidth = 60
	} else {
		HeadingWidth = maxWidth
	}
}

// Pad adds leading spaces to align numbers to 4 characters width.
func Pad(v int) string {
	s := strconv.Itoa(v)
	var b []string
	for i := len(s); i < 4; i++ {
		b = append(b, " ")
	}
	b = append(b, s)
	return strings.Join(b, "")
}

// Adds period to end of string if one is not present.
func AddPeriod(s string) string {
	if strings.LastIndex(s, ".") != len(s)-1 {
		return s + "."
	} else {
		return s
	}
}

// Converts decimal to percent
func ToPercent(f float64) float64 {
	return f * 100
}

// Formats time
func FormatTime(t float64) string {
	return time.Unix(int64(t), 0).Format("15:04")
}

// Limit slice of data, provided the slice is at least the desired length.
func LimitData(d []weather.DataPoint, l int) []weather.DataPoint {
	if len(d) >= l {
		return d[0:l]
	} else {
		return d
	}
}

func MinTemp(f weather.Forecast) {
	fmt.Fprintf(TW, formatValueWithUnit, "Min Temperature", Round(f.Daily.Data[0].TemperatureMin), temperatureUnit)
}

// Prints maximum daily temperature and time.
func MaxTemp(f weather.Forecast) {
	fmt.Fprintf(TW, formatValueWithUnit, "Max Temperature", f.Daily.Data[0].TemperatureMax, temperatureUnit)
	// Show temperature trend if available
	if f.Daily.Data[0].TemperatureTrend != "" {
		fmt.Fprintf(TW, formatLabelValue, "Temp Trend", f.Daily.Data[0].TemperatureTrend, "")
	}
}

// Prints minimum daily temperature and time.
func CurrentTemp(f weather.Forecast) {
	current := f.Hourly.Data[0]
	actual := Round(current.Temperature)
	feelsLike := Round(current.ApparentTemperature)

	// Only show "feels like" if it differs significantly from actual (>3 degrees)
	diff := feelsLike - actual
	if diff > 3 || diff < -3 {
		fmt.Fprintf(TW, "Current Temperature:\t%.0f %s (feels like %.0f %s)\n",
			actual, temperatureUnit, feelsLike, temperatureUnit)
	} else {
		fmt.Fprintf(TW, formatValueWithUnit, "Current Temperature", actual, temperatureUnit)
	}
}

// Prints humidity converted to percent.
func Humidity(f weather.Forecast) {
	fmt.Fprintf(TW, formatValueWithUnit, "Humidity", ToPercent(f.Currently.Humidity), percentUnit)
}

// DegreesToCardinal converts wind bearing in degrees to cardinal direction.
func DegreesToCardinal(degrees float64) string {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	index := int((degrees + 11.25) / 22.5)
	return directions[index%16]
}

// Prints the windspeed average for the day with direction.
func Windspeed(f weather.Forecast) {
	windDir := DegreesToCardinal(f.Currently.WindBearing)
	fmt.Fprintf(TW, "Windspeed:\t%.0f %s from %s\n", f.Currently.WindSpeed, windSpeedUnit, windDir)
	// Show wind gust if available
	if f.Currently.WindGust > 0 {
		fmt.Fprintf(TW, formatValueWithUnit, "Wind Gust", f.Currently.WindGust, windSpeedUnit)
	}
}

// Prints precipitation probability.
func Precipitation(f weather.Forecast) {
	prob := Round(ToPercent(f.Daily.Data[0].PrecipProbability))
	precipType := f.Daily.Data[0].PrecipType

	if precipType != "" && prob > 0 {
		// Capitalize using strings.Title for proper formatting
		fmt.Fprintf(TW, "%s Probability:\t%.0f %s\n", strings.Title(precipType), prob, percentUnit)
	} else {
		fmt.Fprintf(TW, formatValueWithUnit, "Precipitation", prob, percentUnit)
	}
}

// Prints the pressure in atmospheres.
func Pressure(f weather.Forecast) {
	fmt.Fprintf(TW, formatValueWithUnit, "Pressure", f.Daily.Data[0].Pressure, pressureUnit)
}

func Visibility(f weather.Forecast) {
	fmt.Fprintf(TW, formatValueWithUnit, "Visibility", f.Daily.Data[0].Visibility, distanceUnit)
}

// Note: Sunrise/Sunset times are not available from NOAA forecast API.
// Would require astronomical calculations or integration with a separate API like sunrise-sunset.org

// AirQualityIndex takes a forecast and lists the highest AQI index
// and its particle type and category.
func AirQualityIndex(f []air.Forecast) {
	// Check if air forecast data is available
	if len(f) == 0 {
		fmt.Fprintf(TW, formatMultipleValues, "Air Quality Index", "N/A", "No data", "unavailable")
		return
	}

	today := f[0].DateForecast
	aqi := -1 // Initialize to -1 to detect if we found any valid AQI values
	var particle string
	var category string
	var categoryOnly string // For when AQI is unavailable but category is

	for _, measurement := range f {
		// We are only interested in the highest AQI for today.
		if measurement.DateForecast != today {
			break
		}

		// Even if AQI is -1, capture category information
		if measurement.Category.Name != "" && categoryOnly == "" {
			categoryOnly = measurement.Category.Name
		}

		// Skip measurements with invalid AQI values (AirNow returns -1 for unavailable forecasts)
		if measurement.AQI < 0 {
			continue
		}

		// If that measurement exceeds that of the other reigning particle,
		// a new pollutant is crowned.
		if measurement.AQI > aqi {
			aqi = measurement.AQI
			particle = measurement.ParameterName
			category = measurement.Category.Name
		}
	}

	// If we have a valid AQI, show it with details
	if aqi >= 0 {
		fmt.Fprintf(TW, formatMultipleValues, "Air Quality Index", aqi, particle, category)
		return
	}

	// If no AQI but we have category info, show that
	if categoryOnly != "" {
		fmt.Fprintf(TW, formatMultipleValues, "Air Quality", categoryOnly, "(numeric", "forecast pending)")
		return
	}

	// No data at all
	fmt.Fprintf(TW, formatMultipleValues, "Air Quality Index", "N/A", "Forecast", "not yet available")
}

// Prints the summary for the day.
func DailySummary(f weather.Forecast) {
	fmt.Fprintf(TW, formatString, "Currently", AddPeriod(f.Currently.Summary))
	// Show detailed forecast if available
	if f.Currently.DetailedForecast != "" {
		fmt.Fprintf(TW, formatString, "Details", AddPeriod(f.Currently.DetailedForecast))
	}
}

// Prints the summary for the week.
func WeeklySummary(f weather.Forecast) {
	fmt.Fprintf(TW, formatString, "This week", AddPeriod(f.Daily.Summary))
}

// WeatherAlerts prints active weather alerts if any exist.
func WeatherAlerts(f weather.Forecast) {
	if len(f.Alerts) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(Title("Weather Alerts"))
	for i, alert := range f.Alerts {
		if i > 0 {
			fmt.Println()
		}
		fmt.Fprintf(TW, "Alert:\t%s\n", alert.Title)

		// Show expiration time if available
		if alert.Expires > 0 {
			expiryTime := FormatTime(alert.Expires)
			fmt.Fprintf(TW, "Expires:\t%s\n", expiryTime)
		}

		// Show description (truncate if too long)
		description := alert.Description
		if len(description) > 200 {
			description = description[:197] + "..."
		}
		fmt.Fprintf(TW, "Details:\t%s\n", description)
	}
	TW.Flush()
}
