package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"os"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
	"unicode/utf8"
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

// Unit symbols. Output is ASCII only, so temperatures read "46F".
// Symbol units attach to the number ("46F", "83%"); word units take a space ("6 mph").
var temperatureUnit = "F"
var windSpeedUnit = "mph"
var pressureUnit = "inHg"
var distanceUnit = "miles"
var percentUnit = "%"

// TW aligns label/value lists; minwidth keeps short labels from crowding their values.
var TW = tabwriter.NewWriter(output, minwidth, tabwidth, padding, padchar, flags)

// Table aligns multi-column tables. No minwidth, so narrow numeric columns stay narrow
// and an hourly table fits in 80 columns.
var Table = tabwriter.NewWriter(output, 0, tabwidth, padding, padchar, flags)

// MaxReportWidth caps line length so reports stay readable on wide terminals.
const MaxReportWidth = 80

// ReportWidth is the width every section title and wrapped paragraph shares.
var ReportWidth = min(GetTerminalWidth(), MaxReportWidth)

// Format strings for tabwriter output
var formatValueWithUnit = "%s:\t%.0f%s\n"      // e.g., "Humidity: 83%"
var formatValueWithWordUnit = "%s:\t%.0f %s\n" // e.g., "Visibility: 10 miles"
var formatPressure = "%s:\t%.2f %s\n"          // e.g., "Pressure: 30.02 inHg"
var formatString = "%s:\t%s\n"                 // e.g., "Currently: Mostly Cloudy"

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

	if int(retCode) == -1 || ws.Col == 0 {
		// Fallback to 80 columns if detection fails (e.g. output is piped)
		return 80
	}

	return int(ws.Col)
}

// calculateValueColumnWidth calculates the available width for the value column
// in a tabwriter output, accounting for the label column width and padding.
func calculateValueColumnWidth(label string) int {

	// Tabwriter will expand the first column to fit the widest label
	// In our reports, we need to consider common labels
	maxLabelWidth := len(label)
	commonLabels := []string{
		"Current Temperature",
		"Air Quality Index",
		"This Afternoon",
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
	valueWidth := ReportWidth - firstColumnWidth - 5

	// Ensure a reasonable minimum
	if valueWidth < 40 {
		valueWidth = 40
	}

	return valueWidth
}

// Title returns a bold section heading padded with "=" to ReportWidth.
func Title(t string) string {
	title := "== " + strings.ToUpper(t) + " "

	// Calculate remaining space and fill with =
	remaining := ReportWidth - utf8.RuneCountInString(title)
	if remaining > 0 {
		title += strings.Repeat("=", remaining)
	} else {
		title += "=="
	}

	return Bold(title)
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
		fmt.Fprintf(TW, formatString, "Temp Trend", f.Daily.Data[0].TemperatureTrend)
	}
}

// Prints minimum daily temperature and time.
func CurrentTemp(f weather.Forecast) {
	current := GetCurrentHourlyData(f)
	FormatTemperatureWithFeelsLike(TW, "Current Temperature", Round(current.Temperature), Round(current.ApparentTemperature), temperatureUnit)
}

// Prints humidity converted to percent.
func Humidity(f weather.Forecast) {
	fmt.Fprintf(TW, formatValueWithUnit, "Humidity", ToPercent(GetCurrentHourlyData(f).Humidity), percentUnit)
}

// Prints the windspeed average for the day with direction.
func Windspeed(f weather.Forecast) {
	current := GetCurrentHourlyData(f)
	windStr := FormatWindString(current.WindSpeed, current.WindBearing, current.WindGust, windSpeedUnit, true)
	fmt.Fprintf(TW, "Windspeed:\t%s\n", windStr)
}

// Prints precipitation probability.
func Precipitation(f weather.Forecast) {
	prob := Round(ToPercent(f.Daily.Data[0].PrecipProbability))
	precipType := f.Daily.Data[0].PrecipType

	if precipType != "" && prob > 0 {
		// Capitalize using strings.Title for proper formatting
		fmt.Fprintf(TW, "%s Probability:\t%.0f%s\n", CapitalizeFirst(precipType), prob, percentUnit)
	} else {
		fmt.Fprintf(TW, formatValueWithUnit, "Precipitation", prob, percentUnit)
	}
}

// Prints the observed pressure in inches of mercury, when an observation was available.
func Pressure(f weather.Forecast) {
	if p := f.Daily.Data[0].Pressure; p > 0 {
		fmt.Fprintf(TW, formatPressure, "Pressure", p, pressureUnit)
	}
}

// Prints the observed visibility, when an observation was available.
func Visibility(f weather.Forecast) {
	if v := f.Daily.Data[0].Visibility; v > 0 {
		fmt.Fprintf(TW, formatValueWithWordUnit, "Visibility", v, distanceUnit)
	}
}

// Note: Sunrise/Sunset times are not available from NOAA forecast API.
// Would require astronomical calculations or integration with a separate API like sunrise-sunset.org

// AirQualityIndex takes a forecast and lists the highest AQI index
// and its particle type and category.
func AirQualityIndex(f []air.Forecast) {
	// Check if air forecast data is available
	if len(f) == 0 {
		fmt.Fprintf(TW, formatString, "Air Quality", "unavailable")
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
		fmt.Fprintf(TW, "%s:\t%d %s - %s\n", "Air Quality Index", aqi, particle, AQICategory(category))
		return
	}

	// If no AQI but we have category info, show that
	if categoryOnly != "" {
		fmt.Fprintf(TW, formatString, "Air Quality", AQICategory(categoryOnly)+" (numeric forecast pending)")
		return
	}

	// No data at all
	fmt.Fprintf(TW, formatString, "Air Quality", "forecast not yet available")
}

// Prints the summary for the day.
func DailySummary(f weather.Forecast) {
	// Calculate proper width for value column based on terminal size
	maxWidth := calculateValueColumnWidth("Currently")
	wrappedSummary := strings.Join(WrapText(AddPeriod(GetCurrentHourlyData(f).Summary), maxWidth), "\n\t")
	fmt.Fprintf(TW, formatString, "Currently", wrappedSummary)

	// Show detailed forecast if available
	if f.Currently.DetailedForecast != "" {
		wrappedDetails := strings.Join(WrapText(AddPeriod(f.Currently.DetailedForecast), maxWidth), "\n\t")
		fmt.Fprintf(TW, formatString, "Details", wrappedDetails)
	}
}

// PeriodSummary prints NOAA's narrative for the first forecast period, labeled with
// the period's own name ("This Afternoon", "Tonight") so it is never mistaken for the week.
func PeriodSummary(f weather.Forecast) {
	label := "Today"
	if len(f.Daily.Data) > 0 && f.Daily.Data[0].PeriodName != "" {
		label = f.Daily.Data[0].PeriodName
	}
	maxWidth := calculateValueColumnWidth(label)
	wrappedSummary := strings.Join(WrapText(AddPeriod(f.Daily.Summary), maxWidth), "\n\t")
	fmt.Fprintf(TW, formatString, label, wrappedSummary)
}

// WeatherAlerts prints active weather alerts if any exist.
// NWS descriptions are split into their WHAT/WHERE/WHEN/IMPACTS sections when present.
func WeatherAlerts(f weather.Forecast) {
	if len(f.Alerts) == 0 {
		return
	}

	fmt.Println(Title("Weather Alerts"))
	now := time.Now()
	for i, alert := range f.Alerts {
		if i > 0 {
			fmt.Fprintln(TW)
		}

		headline := AlertHeadline(alert.Title)
		if alert.Expires > 0 {
			headline += " | until " + FormatUntil(time.Unix(int64(alert.Expires), 0), now)
		}

		fmt.Fprintf(TW, "Alert:\t%s\n", headline)
		printAlertSections(alert.Description, "Alert:")
	}
	TW.Flush()
	fmt.Println()
}

// printAlertSections writes an NWS alert description to TW as labeled sections
// (What, Where, When, ...), or as one "Details" section when it has no such structure.
// widestOther is the longest other label in the same flush, so wrapping accounts for it.
func printAlertSections(desc, widestOther string) {
	sections := ParseNWSDescription(desc)
	if len(sections) == 0 && desc != "" {
		sections = []LabeledText{{Label: "Details", Text: strings.Join(strings.Fields(desc), " ")}}
	}

	// Wrap to whatever is left after the label column tabwriter will produce.
	labelWidth := len(widestOther)
	for _, sec := range sections {
		labelWidth = max(labelWidth, utf8.RuneCountInString(sec.Label)+1)
	}
	wrapWidth := max(ReportWidth-max(labelWidth+padding, minwidth), 30)

	for _, sec := range sections {
		fmt.Fprintf(TW, "%s:\t%s\n", sec.Label, strings.Join(WrapText(sec.Text, wrapWidth), "\n\t"))
	}
}
