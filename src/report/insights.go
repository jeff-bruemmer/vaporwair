package report

import (
	"fmt"
	"strings"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"text/tabwriter"
)

// conditionsSection prints today's forecast paragraph and current conditions.
func conditionsSection(tw *tabwriter.Writer, w weather.Forecast, a []air.Forecast) {
	daily := w.Daily.Data[0]

	// Name the section after NOAA's period ("This Afternoon", "Tonight") so it's right at any hour.
	fmt.Fprintln(tw, Title(periodLabel(w)))
	summary := daily.DetailedForecast
	if summary == "" {
		summary = daily.Summary
	}
	if summary != "" {
		fmt.Fprintln(tw, strings.Join(WrapText(AddPeriod(summary), ReportWidth), "\n"))
	}
	fmt.Fprintln(tw)

	fmt.Fprintln(tw, Title("Current Conditions"))
	conditionFields(tw, w, a)
}

// futureHours returns up to n hours that start after now. The hour in progress is
// already covered by Current Conditions.
func futureHours(w weather.Forecast, n int) []weather.DataPoint {
	now := float64(clock().Unix())
	for i, h := range w.Hourly.Data {
		if h.Time > now {
			return LimitData(w.Hourly.Data[i:], n)
		}
	}
	return nil
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

	// Next few hours: the hourly report's table, without dewpoint and humidity
	fmt.Println()
	fmt.Println(Title("Next Few Hours"))
	hoursTable(futureHours(w, DefaultMaxHours), false)
	fmt.Println()
}
