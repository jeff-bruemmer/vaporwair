package report

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// AirQuality prints AQI levels for each forecast day and pollutant.
func AirQuality(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Air Quality Forecast"))

	// Category data can exist even when the numeric AQI is still pending (-1)
	hasData := false
	for _, f := range a {
		if f.Category.Name != "" {
			hasData = true
			break
		}
	}
	if !hasData {
		fmt.Println("No air quality data available.")
		return
	}

	// Same layout as the other tables: the date shows once per day, and each day's
	// pollutants are sorted by name so rows line up from one day to the next.
	rows := slices.Clone(a)
	slices.SortStableFunc(rows, func(x, y air.Forecast) int {
		return cmp.Or(
			cmp.Compare(strings.TrimSpace(x.DateForecast), strings.TrimSpace(y.DateForecast)),
			cmp.Compare(x.ParameterName, y.ParameterName))
	})
	printTableHeader([]column{{"Day", ""}, {"Pollutant", ""}, {"AQI", ""}, {"Category", ""}})
	anyPending := false
	prevDate := ""
	for _, f := range rows {
		date := formatAQIDate(f.DateForecast)
		day := date
		if date == prevDate {
			day = ""
		}
		prevDate = date

		aqi := "pending"
		if f.AQI >= 0 {
			aqi = fmt.Sprint(f.AQI)
		} else {
			anyPending = true
		}
		fmt.Fprintf(Table, "%s\t%s\t%s\t%s\n", day, f.ParameterName, aqi, AQICategory(f.Category.Name))
	}
	Table.Flush()

	if anyPending {
		fmt.Println("\nNote: 'pending' AQI values are published later in the day.")
	}
}

// formatAQIDate turns AirNow's "2025-10-24" into "Fri Oct 24", leaving unparseable dates as-is.
func formatAQIDate(d string) string {
	d = strings.TrimSpace(d)
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return d
	}
	return t.Format("Mon Jan 2")
}
