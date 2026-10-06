package report

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// fixtureForecast is a cold, windy, alert-bearing forecast that exercises every report path,
// including sub-zero temperatures that "> 0" guards used to hide.
func fixtureForecast() (weather.Forecast, []air.Forecast) {
	now := time.Now().Truncate(time.Hour)
	var w weather.Forecast
	w.Currently = weather.DataPoint{Time: float64(now.Unix()), Summary: "Light Snow", Pressure: 30.02, Visibility: 4}
	w.Daily.Summary = "Snow likely, mainly after 1pm. High near 12. Wind chill values as low as -15."
	w.Hourly.Summary = "Snow through the evening."

	for i := range 24 {
		w.Hourly.Data = append(w.Hourly.Data, weather.DataPoint{
			Time:                float64(now.Add(time.Duration(i) * time.Hour).Unix()),
			Summary:             "Chance Light Snow",
			Temperature:         float64(10 - i),
			ApparentTemperature: float64(-2 - i),
			DewPoint:            -5,
			PrecipProbability:   0.6,
			PrecipType:          "snow",
			Humidity:            0.8,
			WindSpeed:           18,
			WindGust:            30,
			WindBearing:         315,
		})
	}

	names := []string{"Today", "Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"}
	for i, name := range names {
		w.Daily.Data = append(w.Daily.Data, weather.DataPoint{
			Time:              float64(now.AddDate(0, 0, i).Unix()),
			PeriodName:        name,
			Summary:           "Chance Rain And Snow Showers then Mostly Cloudy and Breezy",
			DetailedForecast:  "Snow likely. Cloudy, with a high near 12.",
			TemperatureMax:    12,
			TemperatureMin:    -5,
			DewPoint:          -8,
			PrecipProbability: 0.6,
			PrecipType:        "snow",
			Humidity:          0.8,
			WindSpeed:         18,
			WindGust:          30,
		})
	}
	w.Daily.Data[0].Pressure = 30.02
	w.Daily.Data[0].Visibility = 4

	w.Alerts = []weather.Alert{{
		Title:       "Winter Storm Warning",
		Description: "* WHAT...Heavy snow. * WHERE...Grafton County. * WHEN...Until 7 PM Saturday.",
		Time:        float64(now.Unix()),
		Expires:     float64(now.Add(20 * time.Hour).Unix()),
	}}

	today := now.Format("2006-01-02")
	a := []air.Forecast{
		{DateForecast: today, ParameterName: "O3", AQI: 42, Category: air.Category{Number: 1, Name: "Good"}},
		{DateForecast: today, ParameterName: "PM2.5", AQI: 155, Category: air.Category{Number: 4, Name: "Unhealthy"}},
		{DateForecast: now.AddDate(0, 0, 1).Format("2006-01-02"), ParameterName: "PM2.5", AQI: -1, Category: air.Category{Number: 2, Name: "Moderate"}},
	}
	return w, a
}

// capture runs a report with stdout and both tabwriters redirected, returning the output.
func capture(t *testing.T, run func()) string {
	t.Helper()
	r, wr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout, oldTW, oldTable, oldColor := os.Stdout, TW, Table, useColor
	os.Stdout = wr
	TW = tabwriter.NewWriter(wr, minwidth, tabwidth, padding, padchar, flags)
	Table = tabwriter.NewWriter(wr, 0, tabwidth, padding, padchar, flags)
	useColor = false
	defer func() { os.Stdout, TW, Table, useColor = oldStdout, oldTW, oldTable, oldColor }()

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()
	run()
	TW.Flush()
	Table.Flush()
	wr.Close()
	return <-done
}

func allReports() map[string]func(weather.Forecast, []air.Forecast) {
	return map[string]func(weather.Forecast, []air.Forecast){
		"hourly":   WeatherHourly,
		"daily":    WeatherDaily,
		"week":     WeatherWeek,
		"alerts":   WeatherAlertsReport,
		"air":      AirQuality,
		"clothing": ClothingReport,
		"insights": InsightsReport,
		"summary":  Summary,
	}
}

// Output is black and white and ASCII only.
func TestReportsAreASCII(t *testing.T) {
	w, a := fixtureForecast()
	for name, report := range allReports() {
		out := capture(t, func() { report(w, a) })
		if out == "" {
			t.Errorf("%s: no output", name)
		}
		for i, line := range strings.Split(out, "\n") {
			for _, r := range line {
				if r > 0x7F || r == 0x1B {
					t.Errorf("%s line %d has non-ASCII or escape %q: %q", name, i+1, r, line)
					break
				}
			}
		}
	}
}

// Tables must fit an 80-column terminal.
func TestTablesFit80Columns(t *testing.T) {
	w, a := fixtureForecast()
	for _, name := range []string{"hourly", "week", "air"} {
		out := capture(t, func() { allReports()[name](w, a) })
		for _, line := range strings.Split(out, "\n") {
			if n := utf8.RuneCountInString(line); n > MaxReportWidth {
				t.Errorf("%s: line is %d columns (> %d): %q", name, n, MaxReportWidth, line)
			}
		}
	}
}

// Zero and sub-zero readings are real values and must be shown.
func TestSubZeroValuesShown(t *testing.T) {
	w, a := fixtureForecast()
	daily := capture(t, func() { WeatherDaily(w, a) })
	for _, want := range []string{"-5F", "-8F"} {
		if !strings.Contains(daily, want) {
			t.Errorf("daily report missing %q:\n%s", want, daily)
		}
	}
	insights := capture(t, func() { InsightsReport(w, a) })
	if !strings.Contains(insights, "Dewpoint:") {
		t.Errorf("insights report hides sub-zero dewpoint:\n%s", insights)
	}
}

func TestAQICategoryMarksUnhealthy(t *testing.T) {
	if got := AQICategory("Good"); got != "Good" {
		t.Errorf("AQICategory(Good) = %q", got)
	}
	if got := AQICategory("Unhealthy for Sensitive Groups"); got != "Unhealthy for Sensitive Groups !" {
		t.Errorf("AQICategory(USG) = %q", got)
	}
}

// Clothing is chosen for the coldest it will feel, not the day's high.
func TestClothingUsesColdestFeelsLike(t *testing.T) {
	w, _ := fixtureForecast()
	rec := GetClothingRecommendation(w)
	if rec.Coldest >= 0 {
		t.Errorf("expected coldest feels-like below zero, got %.0f", rec.Coldest)
	}
	if rec.Outfit != OutfitDescExtremeCold {
		t.Errorf("expected extreme cold outfit for sub-zero feels-like, got %q", rec.Outfit)
	}
}
