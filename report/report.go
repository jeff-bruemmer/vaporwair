package report

import (
	"fmt"
	"text/tabwriter"
	"os"
	"strings"
	"time"
	"github.com/jeff-bruemmer/vaporwair/weather"
	"github.com/jeff-bruemmer/vaporwair/air"
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

// Separator separates report summaries from tables.
var Separator = "+++"

var TW = tabwriter.NewWriter(output, minwidth, tabwidth, padding, padchar, flags)

// Formats
var	f1 = "%s:\t%.0f %s at %v %s\n"
var	f2 = "%s:\t%.0f %s\n"
var	f3 = "%s:\t%v %s\n"

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

// Adds period to end of string.
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

// Formats
var tu = "degrees"
var hm = "HH:MM"
var wu = "mph"
var pu = "atm" 
var du = "miles"
var pc = "%"

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
func Windspeed (f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Windspeed", f.Daily.Data[0].WindSpeed, wu)
}

// Prints the average cloudcover as a percentage.
func Cloudcover (f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Cloudcover", ToPercent(f.Daily.Data[0].CloudCover), pc)
}

// Prints precipitation and type of precipitation.
func Precipitation (f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Precipitation", Round(ToPercent(f.Daily.Data[0].PrecipProbability)), pc)
	if ToPercent(f.Daily.Data[0].PrecipProbability) > 0 {
		fmt.Fprintf(TW, f3, "Precip Type", f.Daily.Data[0].PrecipType, "")
	}
}

// Prints the pressure in atmospheres.
func Pressure (f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Pressure", f.Daily.Data[0].Pressure, pu)
}

func Dewpoint (f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Dewpoint", f.Daily.Data[0].DewPoint, tu)
}

func Visibility(f weather.Forecast) {
	fmt.Fprintf(TW, f2, "Visibility", f.Daily.Data[0].Visibility, du)
}

// Format 3
func AirQualityIndex(f []air.Forecast) {
	fmt.Fprintf(TW, f3, "AQI", f[0].AQI, f[0].Category.Name)
}

func Sunrise(f weather.Forecast) {
	fmt.Fprintf(TW, f3, "Sunrise", FormatTime(f.Daily.Data[0].SunriseTime), hm)
}

func Sunset(f weather.Forecast) {
	fmt.Fprintf(TW, f3, "Sunset", FormatTime(f.Daily.Data[0].SunsetTime), hm)
}

func Today(wf weather.Forecast, af []air.Forecast) {
	CurrentTemp(wf)
	MinTemp(wf)
	MaxTemp(wf)
	AirQualityIndex(af)
	Precipitation(wf)
	Sunrise(wf)
	Sunset(wf)
}
