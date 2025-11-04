package report

import (
	"math"
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
