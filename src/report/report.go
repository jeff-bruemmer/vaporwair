package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"os"
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

// Formats
var tu = "°F"
var hm = "HH:MM"
var wu = "mph"
var pu = "atm"
var du = "miles"
var pc = "%"

// Separator separates report summaries from tables.
var Separator = "+++"

var TW = tabwriter.NewWriter(output, minwidth, tabwidth, padding, padchar, flags)

// Formats
var f1 = "%s:\t%.0f %s at %v %s\n"
var f2 = "%s:\t%.0f %s\n"
var f3 = "%s:\t%v %s\n"
var f4 = "%s:\t%v %s %s\n"
var f5 = "%s:\t%s\n"
var f6 = "%s:\t%v\n"

// Adds title frame
func Title(t string) string {
	return "-- " + strings.ToUpper(t) + " --"
}

// Adds space padding
func Pad(v int) string {
	fmt.Println("v", v)
	s := string(v)
	fmt.Println(s)
	var b []string
	for i := len(s); i < 4; i++ {
		b = append(b, " ")
	}
	b = append(b, s)
	fmt.Println(b)
	return strings.Join(b, "")
}

// Adds period to end of string, because the Dark sky summaries are punctuationally inconsistent.
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

// Format 1
func MinTemp(f weather.Forecast) {
	fmt.Fprintf(TW, f1, "Min Temperature", Round(f.Daily.Data[0].TemperatureMin), tu, FormatTime(f.Daily.Data[0].TemperatureMinTime), hm)
}

// Prints maximum daily temperature and time.
func MaxTemp(f weather.Forecast) {
	fmt.Fprintf(TW, f1, "Max Temperature", f.Daily.Data[0].TemperatureMax, tu, FormatTime(f.Daily.Data[0].TemperatureMaxTime), hm)

}

// Format 2
// Prints minimum daily temperature and time.
func CurrentTemp(f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Current Temperature", Round(f.Hourly.Data[0].Temperature), tu)
}

// Prints humidity converted to percent.
func Humidity(f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Humidity", ToPercent(f.Daily.Data[0].Humidity), pc)
}

// Prints the windspeed average for the day.
func Windspeed(f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Windspeed", f.Currently.WindSpeed, "mph")
}

// Prints the average cloudcover as a percentage.
func Cloudcover(f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Cloudcover", ToPercent(f.Daily.Data[0].CloudCover), pc)
}

// Prints precipitation and type of precipitation.
func Precipitation(f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Precipitation", Round(ToPercent(f.Daily.Data[0].PrecipProbability)), pc)
	if ToPercent(f.Daily.Data[0].PrecipProbability) > 0 {
		fmt.Fprintf(TW, f3, "Precip Type", f.Daily.Data[0].PrecipType, "")
	}
}

// Prints the pressure in atmospheres.
func Pressure(f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Pressure", f.Daily.Data[0].Pressure, pu)
}

func Dewpoint(f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Dewpoint", f.Daily.Data[0].DewPoint, tu)
}

func Visibility(f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Visibility", f.Daily.Data[0].Visibility, du)
}

// Format 3
// Sunrise prints the time the sun rises.
func Sunrise(f weather.Forecast) {
	fmt.Fprintf(TW, f3, "Sunrise", FormatTime(f.Daily.Data[0].SunriseTime), hm)
}

// Sunset prints the time the sun sets.
func Sunset(f weather.Forecast) {
	fmt.Fprintf(TW, f3, "Sunset", FormatTime(f.Daily.Data[0].SunsetTime), hm)
}

// Format 4
// AirQualityIndex takes a forecast and lists the highest AQI index
// and its particle type and category.
func AirQualityIndex(f []air.Forecast) {
	today := f[0].DateForecast
	var aqi int
	var particle string
	var category string
	for _, measurement := range f {
		// We are only interested in the highest AQI for today.
		if measurement.DateForecast != today {
			break
		}

		// If that measurement exceeds that of the other reigning particle,
		// a new pollutant is crowned.
		if measurement.AQI > aqi {
			aqi = measurement.AQI
			particle = measurement.ParameterName
			category = measurement.Category.Name
		}
	}
	fmt.Fprintf(TW, f4, "Air Quality Index", aqi, particle, category)
}

// Format 5
// Prints the summary for the day.
func DailySummary(f weather.Forecast) {
	fmt.Fprintf(TW, f5, "Currently", AddPeriod(f.Currently.Summary))
}

// Prints the summary for the week.
func WeeklySummary(f weather.Forecast) {
	fmt.Fprintf(TW, f5, "This week", AddPeriod(f.Daily.Summary))
}

// Format 6
// Prints the UV index
func UVIndex(f weather.Forecast) {
	fmt.Fprintf(TW, f6, "UV Index", f.Currently.UVIndex)
}
