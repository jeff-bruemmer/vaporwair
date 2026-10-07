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

// Format string for tabwriter label/value rows
var formatString = "%s:\t%s\n" // e.g., "Currently: Mostly Cloudy"

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

// printHighLow prints today's high and low. In the evening NOAA's first period is a night
// with no daytime high (its "high" is just the warmest hour left tonight, about the current
// temperature), so it prints tonight's low and tomorrow's high instead.
func printHighLow(tw *tabwriter.Writer, f weather.Forecast) {
	first := f.Daily.Data[0]
	if !IsNightPeriod(first) {
		fmt.Fprintf(tw, "High / Low:\t%.0f%s / %.0f%s\n", Round(first.TemperatureMax), temperatureUnit, Round(first.TemperatureMin), temperatureUnit)
		return
	}
	fmt.Fprintf(tw, "Low %s:\t%.0f%s\n", first.PeriodName, Round(first.TemperatureMin), temperatureUnit)
	if len(f.Daily.Data) > 1 {
		next := f.Daily.Data[1]
		label := next.PeriodName
		if label == "" {
			label = "Tomorrow"
		}
		fmt.Fprintf(tw, "High %s:\t%.0f%s\n", label, Round(next.TemperatureMax), temperatureUnit)
	}
}

// conditionFields prints Insights' current-conditions list: what you'd decide on now.
// Pressure and visibility are left to the daily report.
func conditionFields(tw *tabwriter.Writer, w weather.Forecast, a []air.Forecast) {
	daily := w.Daily.Data[0]
	current := GetCurrentHourlyData(w)

	FormatTemperatureWithFeelsLike(tw, "Temperature", Round(current.Temperature), Round(current.ApparentTemperature), temperatureUnit)
	printHighLow(tw, w)
	if daily.TemperatureTrend != "" {
		fmt.Fprintf(tw, formatString, "Temperature Trend", daily.TemperatureTrend)
	}

	// A low chance is noise; Next Few Hours shows any chance hour by hour.
	if daily.PrecipProbability >= PrecipSignificantThreshold {
		fmt.Fprintf(tw, "%s Chance:\t%.0f%s\n",
			GetPrecipTypeOrDefault(daily.PrecipType),
			Round(ToPercent(daily.PrecipProbability)),
			percentUnit)
	}

	fmt.Fprintf(tw, "Humidity:\t%.0f%s\n", ToPercent(current.Humidity), percentUnit)
	if current.DewPoint != 0 { // 0 means NOAA sent no dewpoint; see WeatherDaily
		fmt.Fprintf(tw, "Dewpoint:\t%.0f%s\n", current.DewPoint, temperatureUnit)
	}

	windStr := FormatWindString(current.WindSpeed, current.WindBearing, current.WindGust, windSpeedUnit, true)
	fmt.Fprintf(tw, "Wind:\t%s\n", windStr)

	if aqi, ok := airQualityLine(a); ok {
		fmt.Fprintf(tw, formatString, "Air Quality", aqi)
	}
}

// Note: Sunrise/Sunset times are not available from NOAA forecast API.
// Would require astronomical calculations or integration with a separate API like sunrise-sunset.org

// airQualityLine formats today's highest AQI, e.g. "30 AQI (OZONE) - Good".
// It reports false when there is no numeric forecast for today.
func airQualityLine(a []air.Forecast) (string, bool) {
	aqi, particle, category := GetHighestAQIForToday(a)
	if aqi < 0 {
		return "", false
	}
	return fmt.Sprintf("%d AQI (%s) - %s", aqi, particle, AQICategory(category)), true
}

// airQualityStatus explains a missing AQI: a category without a number yet, or no data.
func airQualityStatus(a []air.Forecast) string {
	if len(a) == 0 {
		return "unavailable"
	}
	today := a[0].DateForecast
	for _, measurement := range a {
		if measurement.DateForecast != today {
			break
		}
		if measurement.Category.Name != "" {
			return AQICategory(measurement.Category.Name) + " (numeric forecast pending)"
		}
	}
	return "forecast not yet available"
}

// WeatherAlerts prints a short block for each active alert: the headline with when it
// applies, and what it is. The alerts report has the full NWS text.
func WeatherAlerts(f weather.Forecast) {
	if len(f.Alerts) == 0 {
		return
	}

	fmt.Println(Title("Weather Alerts"))
	now := clock()
	for i, alert := range f.Alerts {
		if i > 0 {
			fmt.Fprintln(TW)
		}

		headline := AlertHeadline(alert.Title)
		if window := FormatAlertWindow(unixOrZero(alert.Time), unixOrZero(alert.Expires), now); window != "" {
			headline += " | " + window
		}

		fmt.Fprintf(TW, "Alert:\t%s\n", headline)
		if what := alertWhat(alert.Description); what != "" {
			fmt.Fprintf(TW, "What:\t%s\n", strings.Join(WrapText(what, alertWrapWidth("Alert:")), "\n\t"))
		}
	}
	fmt.Fprintf(TW, "Details:\tvaporwair alerts\n")
	TW.Flush()
	fmt.Println()
}

// alertWhat returns an alert's WHAT section, or failing that its first sentence.
func alertWhat(desc string) string {
	for _, sec := range ParseNWSDescription(desc) {
		if sec.Label == "What" {
			return sec.Text
		}
	}
	text := strings.Join(strings.Fields(desc), " ")
	if i := strings.Index(text, ". "); i >= 0 {
		text = text[:i+1]
	}
	return Truncate(text, 160)
}

// alertWrapWidth is the room left for alert text after a label column as wide as widest.
func alertWrapWidth(widest string) int {
	return max(ReportWidth-max(len(widest)+padding, minwidth), 30)
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
	wrapWidth := alertWrapWidth(strings.Repeat(" ", labelWidth))

	for _, sec := range sections {
		fmt.Fprintf(TW, "%s:\t%s\n", sec.Label, strings.Join(WrapText(sec.Text, wrapWidth), "\n\t"))
	}
}
