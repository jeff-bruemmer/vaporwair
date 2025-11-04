package report

import (
	"bytes"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"os"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
	"unsafe"
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

// HeadingWidth is set dynamically based on terminal width or content
var HeadingWidth = GetTerminalWidth()

// Format strings for tabwriter output
var formatValueWithTime = "%s:\t%.0f %s at %v %s\n" // e.g., "Min Temperature: 33 °F at 19:00 HH:MM"
var formatValueWithUnit = "%s:\t%.0f %s\n"          // e.g., "Humidity: 83 %"
var formatLabelValue = "%s:\t%v %s\n"               // e.g., "Sunrise: 06:15 HH:MM"
var formatMultipleValues = "%s:\t%v %s %s\n"        // e.g., "Air Quality Index: 55 O3 Moderate"
var formatString = "%s:\t%s\n"                      // e.g., "Currently: Mostly Cloudy"
var formatNumber = "%s:\t%v\n"                      // e.g., "UV Index: 0"

// winsize struct for terminal size detection
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// GetTerminalWidth detects the current terminal width.
// Falls back to 80 columns if detection fails.
func GetTerminalWidth() int {
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		// Fallback to 80 columns if detection fails
		return 80
	}

	return int(ws.Col)
}

// wrapTextForTabwriter wraps text to fit within the terminal width while preserving words.
// Returns a single string with tab-indented line breaks for use with tabwriter.
func wrapTextForTabwriter(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLength := 0

	for i, word := range words {
		wordLen := len(word)

		// If adding this word would exceed the width, start a new line
		if lineLength+wordLen > maxWidth && lineLength > 0 {
			result.WriteString("\n\t")
			lineLength = 0
		} else if i > 0 && lineLength > 0 {
			result.WriteString(" ")
			lineLength++
		}

		result.WriteString(word)
		lineLength += wordLen
	}

	return result.String()
}

// calculateValueColumnWidth calculates the available width for the value column
// in a tabwriter output, accounting for the label column width and padding.
func calculateValueColumnWidth(label string) int {
	termWidth := GetTerminalWidth()

	// Tabwriter will expand the first column to fit the widest label
	// In our reports, we need to consider common labels
	maxLabelWidth := len(label)
	commonLabels := []string{
		"Current Temperature",
		"Air Quality Index",
		"Tomorrow vs Today",
		"This week",
	}
	for _, l := range commonLabels {
		if len(l) > maxLabelWidth {
			maxLabelWidth = len(l)
		}
	}

	// Ensure minimum width from tabwriter config
	if maxLabelWidth < minwidth {
		maxLabelWidth = minwidth
	}

	// Account for tabwriter padding
	firstColumnWidth := maxLabelWidth + padding

	// Calculate available width for value column
	// Leave some margin for safety
	valueWidth := termWidth - firstColumnWidth - 5

	// Ensure a reasonable minimum
	if valueWidth < 40 {
		valueWidth = 40
	}

	return valueWidth
}

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

	// Get terminal width to ensure we don't exceed it
	termWidth := GetTerminalWidth()

	// Set heading width (with a minimum of 60, maximum of terminal width)
	if maxWidth < 60 {
		HeadingWidth = 60
	} else if maxWidth > termWidth {
		HeadingWidth = termWidth
	} else {
		HeadingWidth = maxWidth
	}
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
	FormatTemperatureWithFeelsLike(TW, "Current Temperature", Round(current.Temperature), Round(current.ApparentTemperature), temperatureUnit)
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

	aqi, particle, category := GetHighestAQIForToday(f)

	// Capture category information even if AQI is unavailable
	var categoryOnly string
	if aqi < 0 {
		today := f[0].DateForecast
		for _, measurement := range f {
			if measurement.DateForecast != today {
				break
			}
			if measurement.Category.Name != "" {
				categoryOnly = measurement.Category.Name
				break
			}
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
	// Calculate proper width for value column based on terminal size
	maxWidth := calculateValueColumnWidth("Currently")
	wrappedSummary := wrapTextForTabwriter(AddPeriod(f.Currently.Summary), maxWidth)
	fmt.Fprintf(TW, formatString, "Currently", wrappedSummary)

	// Show detailed forecast if available
	if f.Currently.DetailedForecast != "" {
		wrappedDetails := wrapTextForTabwriter(AddPeriod(f.Currently.DetailedForecast), maxWidth)
		fmt.Fprintf(TW, formatString, "Details", wrappedDetails)
	}
}

// Prints the summary for the week.
func WeeklySummary(f weather.Forecast) {
	// Calculate proper width for value column based on terminal size
	maxWidth := calculateValueColumnWidth("This week")
	wrappedSummary := wrapTextForTabwriter(AddPeriod(f.Daily.Summary), maxWidth)
	fmt.Fprintf(TW, formatString, "This week", wrappedSummary)
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

		// Wrap description instead of truncating
		maxWidth := calculateValueColumnWidth("Details")
		wrappedDescription := wrapTextForTabwriter(alert.Description, maxWidth)
		fmt.Fprintf(TW, "Details:\t%s\n", wrappedDescription)
	}
	TW.Flush()
}
