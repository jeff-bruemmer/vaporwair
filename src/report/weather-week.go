package report

import (
	"fmt"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// weekFixedWidth approximates the columns before Conditions (a long day name like
// "This Afternoon" plus five numeric columns and padding), so Conditions fits what's left.
const weekFixedWidth = 47

func WeatherWeek(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Week Ahead"))
	// No humidity column: NOAA's daily periods don't include it.
	fmt.Fprintf(Table, "Day\tLow\tHigh\tPrecip\tWind\tGust\tConditions\n")
	fmt.Fprintf(Table, "\t%s\t%s\t%s\t%s\t%s\t\n",
		temperatureUnit, temperatureUnit, percentUnit, windSpeedUnit, windSpeedUnit)
	for _, day := range LimitData(w.Daily.Data, 7) {
		gustStr := "-"
		if day.WindGust > 0 {
			gustStr = fmt.Sprintf("%.0f", day.WindGust)
		}
		// Period name (e.g., "Today", "Monday") is clearer than abbreviated weekday
		dayLabel := day.PeriodName
		if dayLabel == "" {
			dayLabel = time.Unix(int64(day.Time), 0).Format("Mon")
		}
		fmt.Fprintf(Table, "%s\t%.0f\t%.0f\t%.0f\t%.0f\t%s\t%s\n",
			dayLabel,
			day.TemperatureMin,
			day.TemperatureMax,
			ToPercent(day.PrecipProbability),
			day.WindSpeed,
			gustStr,
			Truncate(day.Summary, max(ReportWidth-weekFixedWidth, 12)),
		)
	}
	Table.Flush()
}
