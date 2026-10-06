package report

import (
	"fmt"
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

	// Fixed-width columns so every date's rows line up under one header.
	// Pollutant names from AirNow are short (O3, OZONE, PM2.5, PM10, NO2, CO).
	row := "  %-9s  %7s  %s\n"
	anyPending := false
	date := ""
	fmt.Printf(row, "Pollutant", "AQI", "Category")
	fmt.Printf(row, "---------", "---", "--------")
	for _, f := range a {
		if f.DateForecast != date {
			date = f.DateForecast
			fmt.Println(formatAQIDate(date))
		}

		aqi := "pending"
		if f.AQI >= 0 {
			aqi = fmt.Sprint(f.AQI)
		} else {
			anyPending = true
		}
		fmt.Printf(row, f.ParameterName, aqi, AQICategory(f.Category.Name))
	}

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
