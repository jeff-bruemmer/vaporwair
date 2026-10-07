package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// WeatherDaily displays comprehensive daily forecast with all available NOAA fields.
// There is no overall summary on top: NOAA's is the first day's forecast, shown below.
func WeatherDaily(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Daily Forecast"))

	data := LimitData(w.Daily.Data, 7)

	// Collect every day's fields first, so all days share one label column.
	// tabwriter sizes each flush separately, which made the columns shift from day to day.
	days := make([][]LabeledText, len(data))
	labelWidth := minwidth
	for i, day := range data {
		days[i] = dailyFields(day)
		for _, f := range days[i] {
			labelWidth = max(labelWidth, len(f.Label)+1+padding)
		}
	}
	wrapWidth := max(ReportWidth-labelWidth, 30)
	indent := strings.Repeat(" ", labelWidth)

	for i, day := range data {
		if i > 0 {
			fmt.Println()
		}

		dayLabel := day.PeriodName
		if dayLabel == "" {
			dayLabel = time.Unix(int64(day.Time), 0).Format("Monday, Jan 2")
		}
		fmt.Println(Bold(dayLabel))

		for _, f := range days[i] {
			lines := WrapText(f.Text, wrapWidth)
			fmt.Printf("%-*s%s\n", labelWidth, f.Label+":", strings.Join(lines, "\n"+indent))
		}
	}
}

// dailyFields returns one day's labeled values, in display order.
func dailyFields(day weather.DataPoint) []LabeledText {
	var fields []LabeledText
	add := func(label, format string, args ...any) {
		fields = append(fields, LabeledText{Label: label, Text: fmt.Sprintf(format, args...)})
	}

	// Temperature information (0F and below are real readings, so no "> 0" guards).
	// A night period has no daytime high; NOAA's value is just the warmest hour left.
	if !IsNightPeriod(day) {
		add("High", "%.0f%s", day.TemperatureMax, temperatureUnit)
	}
	add("Low", "%.0f%s", day.TemperatureMin, temperatureUnit)
	if day.TemperatureTrend != "" {
		add("Temperature Trend", "%s", day.TemperatureTrend)
	}

	if day.PrecipProbability > 0 {
		add(GetPrecipTypeOrDefault(day.PrecipType)+" Chance", "%.0f%s", ToPercent(day.PrecipProbability), percentUnit)
	}

	if day.WindSpeed > 0 {
		add("Wind", "%s", FormatWindString(day.WindSpeed, day.WindBearing, day.WindGust, windSpeedUnit, true))
	}

	if day.Humidity > 0 {
		add("Humidity", "%.0f%s", ToPercent(day.Humidity), percentUnit)
	}
	// NOAA's daily periods usually omit dewpoint (stored as 0). A converted Celsius
	// reading is essentially never exactly 0F, so 0 here means missing.
	if day.DewPoint != 0 {
		add("Dewpoint", "%.0f%s", day.DewPoint, temperatureUnit)
	}

	// Pressure and visibility come from a live observation, so only today has them
	if day.Pressure > 0 {
		add("Pressure", "%.2f %s", day.Pressure, pressureUnit)
	}
	if day.Visibility > 0 {
		add("Visibility", "%.0f %s", day.Visibility, distanceUnit)
	}

	if day.DetailedForecast != "" {
		add("Forecast", "%s", AddPeriod(day.DetailedForecast))
	} else if day.Summary != "" {
		add("Summary", "%s", AddPeriod(day.Summary))
	}
	return fields
}
