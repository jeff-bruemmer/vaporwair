package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

func WeatherHourly(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Hourly Summary"))
	hoursTable(LimitData(upcomingHours(w.Hourly.Data), DefaultHourlyLimit), true)
}

// hoursTable prints one row per hour. Insights' Next Few Hours and the hourly report share it,
// so both use the same columns in the same order; detail adds dewpoint and humidity.
// Precip and Gust columns appear only when some hour has a value.
func hoursTable(hours []weather.DataPoint, detail bool) {
	showPrecip, showGust := anyPrecip(hours), anyGust(hours)

	// Units live in the second header row so cells hold bare numbers and the table fits 80 columns.
	cols := []column{{"Time", ""}, {"Temp", temperatureUnit}, {"Feels", temperatureUnit}}
	if detail {
		cols = append(cols, column{"Dew", temperatureUnit})
	}
	if showPrecip {
		cols = append(cols, column{"Precip", percentUnit})
	}
	if detail {
		cols = append(cols, column{"Humid", percentUnit})
	}
	cols = append(cols, column{"Wind", windSpeedUnit})
	if showGust {
		cols = append(cols, column{"Gust", windSpeedUnit})
	}
	fixed := tableFixedWidth(cols, len("00:00"))
	cols = append(cols, column{"Conditions", ""})
	printTableHeader(cols)

	// Leave conditions blank when unchanged from the row above, so changes stand out.
	prevConditions := ""
	for _, h := range hours {
		conditions := Truncate(h.Summary, max(ReportWidth-fixed, 12))
		if h.Summary == prevConditions {
			conditions = ""
		}
		prevConditions = h.Summary

		fmt.Fprintf(Table, "%s\t%.0f\t%.0f\t", FormatTime(h.Time), Round(h.Temperature), Round(h.ApparentTemperature))
		if detail {
			fmt.Fprintf(Table, "%.0f\t", h.DewPoint)
		}
		if showPrecip {
			fmt.Fprintf(Table, "%.0f\t", ToPercent(h.PrecipProbability))
		}
		if detail {
			fmt.Fprintf(Table, "%.0f\t", ToPercent(h.Humidity))
		}
		fmt.Fprintf(Table, "%.0f %s\t", h.WindSpeed, DegreesToCardinal(h.WindBearing))
		if showGust {
			fmt.Fprintf(Table, "%s\t", gustCell(h.WindGust))
		}
		fmt.Fprintf(Table, "%s\n", conditions)
	}
	Table.Flush()
}

// anyPrecip reports whether any data point has a chance of precipitation,
// so an all-zero Precip column can be dropped.
func anyPrecip(data []weather.DataPoint) bool {
	for _, d := range data {
		if d.PrecipProbability > 0 {
			return true
		}
	}
	return false
}
