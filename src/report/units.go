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

// Round rounds a float64 to the nearest integer value
func Round(v float64) float64 {
	return math.Round(v)
}
