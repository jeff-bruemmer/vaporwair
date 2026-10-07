package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// weekDayWidth is a long day name like "This Afternoon", for sizing the Conditions column.
const weekDayWidth = len("This Afternoon")

func WeatherWeek(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Week Ahead"))
	days := LimitData(w.Daily.Data, 7)
	showGust := anyGust(days)

	// No humidity column: NOAA's daily periods don't include it.
	cols := []column{{"Day", ""}, {"Low", temperatureUnit}, {"High", temperatureUnit}, {"Precip", percentUnit}, {"Wind", windSpeedUnit}}
	if showGust {
		cols = append(cols, column{"Gust", windSpeedUnit})
	}
	fixed := tableFixedWidth(cols, weekDayWidth)
	cols = append(cols, column{"Conditions", ""})
	printTableHeader(cols)

	for _, day := range days {
		// Period name (e.g., "Today", "Monday") is clearer than abbreviated weekday
		dayLabel := day.PeriodName
		if dayLabel == "" {
			dayLabel = time.Unix(int64(day.Time), 0).Format("Mon")
		}
		// A night period has no daytime high; NOAA's value is just the warmest hour left.
		high := "-"
		if !IsNightPeriod(day.PeriodName) {
			high = fmt.Sprintf("%.0f", day.TemperatureMax)
		}
		fmt.Fprintf(Table, "%s\t%.0f\t%s\t%.0f\t%.0f %s\t",
			dayLabel,
			day.TemperatureMin,
			high,
			ToPercent(day.PrecipProbability),
			day.WindSpeed,
			DegreesToCardinal(day.WindBearing),
		)
		if showGust {
			fmt.Fprintf(Table, "%s\t", gustCell(day.WindGust))
		}
		fmt.Fprintf(Table, "%s\n", Truncate(day.Summary, max(ReportWidth-fixed, 12)))
	}
	Table.Flush()
}

// column is one table column: its header and the unit shown beneath it.
type column struct {
	name, unit string
}

// printTableHeader writes the header row and, under it, the units row.
// Tables without units (air quality, the outfit scale) get no units row.
func printTableHeader(cols []column) {
	names := make([]string, len(cols))
	units := make([]string, len(cols))
	hasUnits := false
	for i, c := range cols {
		names[i], units[i] = c.name, c.unit
		hasUnits = hasUnits || c.unit != ""
	}
	fmt.Fprintln(Table, strings.Join(names, "\t"))
	if hasUnits {
		fmt.Fprintln(Table, strings.Join(units, "\t"))
	}
}

// tableFixedWidth estimates the width of cols as Table will lay them out, given the widest
// first-column cell, so a trailing Conditions column can be truncated to what's left.
func tableFixedWidth(cols []column, firstWidth int) int {
	width := max(firstWidth, len(cols[0].name)) + padding
	for _, c := range cols[1:] {
		cell := max(len(c.name), len(c.unit), 3) // numbers are at most 3 wide ("-15", "100")
		if c.name == "Wind" {
			cell = max(cell, len("18 NNW")) // wind cells carry a direction
		}
		width += cell + padding
	}
	return width
}

// anyGust reports whether any data point has a gust, so an all-blank Gust column can be dropped.
func anyGust(data []weather.DataPoint) bool {
	for _, d := range data {
		if d.WindGust > 0 {
			return true
		}
	}
	return false
}

// gustCell formats a gust speed, or "-" for an hour or day without one.
func gustCell(g float64) string {
	if g > 0 {
		return fmt.Sprintf("%.0f", g)
	}
	return "-"
}
