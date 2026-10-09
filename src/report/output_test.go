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
	for _, name := range []string{"hourly", "week", "daily", "air"} {
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
	rec := GetClothingRecommendation(w, nil)
	if rec.Coldest >= 0 {
		t.Errorf("expected coldest feels-like below zero, got %.0f", rec.Coldest)
	}
	if rec.Outfit != OutfitDescExtremeCold {
		t.Errorf("expected extreme cold outfit for sub-zero feels-like, got %q", rec.Outfit)
	}
}

// Gust columns appear only when some row has a gust.
func TestGustColumnOnlyWithGusts(t *testing.T) {
	w, a := fixtureForecast()
	for _, name := range []string{"hourly", "week"} {
		if out := capture(t, func() { allReports()[name](w, a) }); !strings.Contains(out, "Gust") {
			t.Errorf("%s: fixture has gusts but no Gust column:\n%s", name, out)
		}
	}
	for i := range w.Hourly.Data {
		w.Hourly.Data[i].WindGust = 0
	}
	for i := range w.Daily.Data {
		w.Daily.Data[i].WindGust = 0
	}
	for _, name := range []string{"hourly", "week"} {
		if out := capture(t, func() { allReports()[name](w, a) }); strings.Contains(out, "Gust") {
			t.Errorf("%s: no gusts but a Gust column:\n%s", name, out)
		}
	}
}

func TestHourlyHasConditions(t *testing.T) {
	w, a := fixtureForecast()
	out := capture(t, func() { WeatherHourly(w, a) })
	if !strings.Contains(out, "Conditions") || strings.Count(out, "Chance Light Snow") != 1 {
		t.Errorf("want a Conditions column showing an unchanged summary once:\n%s", out)
	}
}

// The daily report shows each narrative once, and every day's values start in the same column.
func TestDailyAlignedWithoutRepeats(t *testing.T) {
	w, a := fixtureForecast()
	w.Daily.Data[0].DetailedForecast = "Snow likely, mainly after 1pm. High near 12. Wind chill values as low as -15."
	w.Daily.Data[3].PrecipType = "" // a longer "Precipitation Chance" label on one day
	out := capture(t, func() { WeatherDaily(w, a) })
	if n := strings.Count(out, "Snow likely, mainly after 1pm"); n != 1 {
		t.Errorf("first day's forecast appears %d times:\n%s", n, out)
	}
	col := -1
	for _, line := range strings.Split(out, "\n") {
		label, rest, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(line, " ") {
			continue
		}
		start := len(label) + 1 + len(rest) - len(strings.TrimLeft(rest, " "))
		if col == -1 {
			col = start
		} else if start != col {
			t.Errorf("value starts at column %d, want %d: %q", start, col, line)
		}
	}
}

// Insights uses the standard labels and leaves pressure and visibility to the daily report.
func TestInsightsLabels(t *testing.T) {
	w, a := fixtureForecast()
	out := capture(t, func() { InsightsReport(w, a) })
	for _, want := range []string{"Temperature:", "Wind:", "Snow Chance:", "Air Quality:  ", "155 AQI (PM2.5) - Unhealthy !"} {
		if !strings.Contains(out, want) {
			t.Errorf("insights missing %q:\n%s", want, out)
		}
	}
	for _, bad := range []string{"Windspeed", "Air Quality Index", "Current Temperature", "Probability", "Pressure:", "Visibility:"} {
		if strings.Contains(out, bad) {
			t.Errorf("insights has %q:\n%s", bad, out)
		}
	}
}

// A low precipitation chance is noise in Insights' list.
func TestInsightsHidesLowPrecipChance(t *testing.T) {
	w, a := fixtureForecast()
	w.Daily.Data[0].PrecipProbability = 0.01
	if out := capture(t, func() { InsightsReport(w, a) }); strings.Contains(out, "Snow Chance:") {
		t.Errorf("insights shows a 1%% chance:\n%s", out)
	}
}

// Summary is a glance: a few labeled lines with the air quality and alert headline.
func TestSummaryIsGlance(t *testing.T) {
	w, a := fixtureForecast()
	out := strings.TrimSpace(capture(t, func() { Summary(w, a) }))
	if n := len(strings.Split(out, "\n")); n > 7 {
		t.Errorf("summary is %d lines, want at most 7:\n%s", n, out)
	}
	for _, want := range []string{"Now:", "10F (feels -2F)", "High / Low:", "Outfit:", "Air Quality:  155 AQI (PM2.5) - Unhealthy !", "! WINTER STORM WARNING"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q:\n%s", want, out)
		}
	}
}

// Pointer lines to fuller reports align with the label column above them.
func TestPointersAligned(t *testing.T) {
	w, a := fixtureForecast()
	out := capture(t, func() { InsightsReport(w, a) })
	for _, want := range []string{"Details:  vaporwair alerts", "More:     vaporwair clothing"} {
		if !strings.Contains(out, want) {
			t.Errorf("insights missing aligned %q:\n%s", want, out)
		}
	}
}

// Next Few Hours uses the hourly report's columns, with Conditions last.
func TestNextHoursMatchesHourlyColumns(t *testing.T) {
	w, a := fixtureForecast()
	out := capture(t, func() { InsightsReport(w, a) })
	if !strings.Contains(out, "Time   Temp  Feels  Precip  Wind   Gust  Conditions") {
		t.Errorf("unexpected Next Few Hours header:\n%s", out)
	}
}

// Each day's pollutants are sorted, so rows line up from day to day.
func TestAirPollutantsSorted(t *testing.T) {
	w, _ := fixtureForecast()
	today := time.Now().Format("2006-01-02")
	a := []air.Forecast{
		{DateForecast: today, ParameterName: "PM2.5", AQI: 25, Category: air.Category{Name: "Good"}},
		{DateForecast: today, ParameterName: "OZONE", AQI: 30, Category: air.Category{Name: "Good"}},
	}
	out := capture(t, func() { AirQuality(w, a) })
	if strings.Index(out, "OZONE") > strings.Index(out, "PM2.5") {
		t.Errorf("pollutants not sorted:\n%s", out)
	}
	if strings.Contains(out, "---") {
		t.Errorf("air report still uses dashed headers:\n%s", out)
	}
}

// Insights shows a short alert; the alerts report has the full text.
func TestInsightsAlertIsShort(t *testing.T) {
	w, a := fixtureForecast()
	out := capture(t, func() { InsightsReport(w, a) })
	if !strings.Contains(out, "What:") || strings.Contains(out, "Where:") {
		t.Errorf("want What only:\n%s", out)
	}
	full := capture(t, func() { WeatherAlertsReport(w, a) })
	if !strings.Contains(full, "Where:") {
		t.Errorf("alerts report lost the Where section:\n%s", full)
	}
}
