package report

import (
	"fmt"
	"strings"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"text/tabwriter"
	"time"
)

// conditionsSection prints today's forecast paragraph and current conditions.
func conditionsSection(tw *tabwriter.Writer, w weather.Forecast, a []air.Forecast) {
	daily := w.Daily.Data[0]

	// Find the most current hourly data point (closest to now)
	current := GetCurrentHourlyData(w)

	fmt.Fprintln(tw, Title("Today's Forecast"))
	summary := daily.DetailedForecast
	if summary == "" {
		summary = daily.Summary
	}
	if summary != "" {
		fmt.Fprintln(tw, strings.Join(WrapText(AddPeriod(summary), ReportWidth), "\n"))
	}
	fmt.Fprintln(tw)

	fmt.Fprintln(tw, Title("Current Conditions"))

	// Temperature with feels like
	FormatTemperatureWithFeelsLike(tw, "Temperature", Round(current.Temperature), Round(current.ApparentTemperature), temperatureUnit)
	fmt.Fprintf(tw, "High / Low:\t%.0f%s / %.0f%s\n", daily.TemperatureMax, temperatureUnit, Round(daily.TemperatureMin), temperatureUnit)

	// Precipitation
	if daily.PrecipProbability > 0 {
		fmt.Fprintf(tw, "%s Chance:\t%.0f%s\n",
			GetPrecipTypeOrDefault(daily.PrecipType),
			Round(ToPercent(daily.PrecipProbability)),
			percentUnit)
	}

	// Humidity and dewpoint
	fmt.Fprintf(tw, "Humidity:\t%.0f%s\n", ToPercent(current.Humidity), percentUnit)
	if current.DewPoint > 0 {
		fmt.Fprintf(tw, "Dewpoint:\t%.0f%s\n", current.DewPoint, temperatureUnit)
	}

	// Wind
	windStr := FormatWindString(current.WindSpeed, current.WindBearing, current.WindGust, windSpeedUnit, true)
	fmt.Fprintf(tw, "Wind:\t%s\n", windStr)

	// Pressure and visibility
	if current.Pressure > 0 {
		fmt.Fprintf(tw, formatValueWithUnit, "Pressure", current.Pressure, pressureUnit)
	}
	if current.Visibility > 0 {
		fmt.Fprintf(tw, formatValueWithUnit, "Visibility", current.Visibility, distanceUnit)
	}

	// Air Quality
	if len(a) > 0 {
		aqi, particle, category := GetHighestAQIForToday(a)
		if aqi >= 0 {
			fmt.Fprintf(tw, "Air Quality:\t%d AQI (%s) - %s\n", aqi, particle, category)
		}
	}
}

// hourRow holds the formatted cells for one row of the Next Few Hours table.
type hourRow struct {
	time, temp, conditions, precip, wind string
	hasPrecip                            bool
}

// nextHours returns formatted rows for up to maxHours future hours.
func nextHours(w weather.Forecast, maxHours int) []hourRow {
	currentTime := float64(time.Now().Unix())
	var rows []hourRow

	for _, hour := range w.Hourly.Data {
		if len(rows) >= maxHours {
			break
		}
		// Only show hours that are in the future
		if hour.Time <= currentTime {
			continue
		}

		row := hourRow{
			time:       hour.PeriodName,
			temp:       FormatTemperatureString(Round(hour.Temperature), Round(hour.ApparentTemperature), temperatureUnit),
			conditions: hour.Summary,
			precip:     "0%",
			wind:       FormatWindString(hour.WindSpeed, hour.WindBearing, hour.WindGust, windSpeedUnit, false),
		}
		if row.time == "" {
			row.time = FormatTime(hour.Time)
		}
		if hour.PrecipProbability > 0 {
			row.hasPrecip = true
			kind := hour.PrecipType
			if kind == "" {
				kind = "precip"
			}
			row.precip = fmt.Sprintf("%.0f%% %s", ToPercent(hour.PrecipProbability), kind)
		}
		rows = append(rows, row)
	}
	return rows
}

// InsightsReport provides today's forecast, what to wear, and next few hours.
func InsightsReport(w weather.Forecast, a []air.Forecast) {
	// Show weather alerts first if any exist
	WeatherAlerts(w)

	conditionsSection(TW, w, a)
	TW.Flush()

	// What to Wair
	fmt.Println()
	ClothingSummary(w, a)

	// Next few hours
	fmt.Println()
	fmt.Println(Title("Next Few Hours"))

	rows := nextHours(w, DefaultMaxHours)

	// Drop the precipitation column when it would be all zeros.
	showPrecip := false
	for _, r := range rows {
		showPrecip = showPrecip || r.hasPrecip
	}

	if showPrecip {
		fmt.Fprintf(TW, "Time\tTemp\tConditions\tPrecip\tWind\n")
	} else {
		fmt.Fprintf(TW, "Time\tTemp\tConditions\tWind\n")
	}

	// Leave conditions and wind blank when unchanged from the row above, so changes stand out.
	prev := hourRow{}
	for _, r := range rows {
		conditions, wind := r.conditions, r.wind
		if conditions == prev.conditions {
			conditions = ""
		}
		if wind == prev.wind {
			wind = ""
		}
		if showPrecip {
			fmt.Fprintf(TW, "%s\t%s\t%s\t%s\t%s\n", r.time, r.temp, conditions, r.precip, wind)
		} else {
			fmt.Fprintf(TW, "%s\t%s\t%s\t%s\n", r.time, r.temp, conditions, wind)
		}
		prev = r
	}

	TW.Flush()
	fmt.Println()
}
