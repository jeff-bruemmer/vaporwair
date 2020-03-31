package report

import (
	"math"
	"strings"
)

/*
units=[units] optional
Return weather conditions in the requested units. [units] should be one of the following:
summary: Any summaries containing temperature or snow accumulation units will have their values in degrees Celsius or in centimeters (respectively).
nearestStormDistance: Kilometers.
precipIntensity: Millimeters per hour.
precipIntensityMax: Millimeters per hour.
precipAccumulation: Centimeters.
temperature: Degrees Celsius.
temperatureMin: Degrees Celsius.
temperatureMax: Degrees Celsius.
apparentTemperature: Degrees Celsius.
dewPoint: Degrees Celsius.
windSpeed: Meters per second.
pressure: Hectopascals.
visibility: Kilometers.
*/

// Selects between US and Standard International (metric) units.
func selectUnit(si, us string) func(string) string {
	return func(s string) string {
		switch strings.ToLower(s) {
		case "us":
			return us

		default:
			return si
		}
	}
}

var WindSpeedUnit = selectUnit("m/s", "mph")
var TemperatureUnit = selectUnit("\u00B0C", "\u00B0F")
var DistanceUnit = selectUnit("km", "mi.")
var VisibilityUnit = DistanceUnit
var NearestStormDistanceUnit = DistanceUnit
var PrecipIntensityUnit = selectUnit("mm/h", "in/h")
var PressureUnit = selectUnit("atm", "atm") // selectUnit("hPa", "atm")

var precision = 0


func digits(p float64) func(float64) float64 {
	return func(v float64) float64 {
		var rounded float64
		pow := math.Pow(10, p)
		d := pow * v
		_, div := math.Modf(d)
		if div >= 0.5 {
			rounded = math.Ceil(d)
		} else {
			rounded = math.Floor(d)
		}
		return rounded / pow
	}
}

var Round = digits(0)
