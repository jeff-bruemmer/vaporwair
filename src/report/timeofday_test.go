package report

import (
	"strings"
	"testing"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// at returns Tue Oct 6 2026 at hh:mm local time.
func at(hh, mm int) time.Time {
	return time.Date(2026, time.October, 6, hh, mm, 0, 0, time.Local)
}

// withClock pins the report clock to t for the rest of the test.
func withClock(t *testing.T, now time.Time) {
	t.Helper()
	old := clock
	clock = func() time.Time { return now }
	t.Cleanup(func() { clock = old })
}

// eveningForecast is a clear autumn evening at 19:11 that cools from 51F to a pre-dawn
// feels-like of 30F at 06:00, with tomorrow's high of 61F.
func eveningForecast() weather.Forecast {
	start := at(19, 0)
	var w weather.Forecast
	feels := []float64{51, 46, 45, 43, 42, 42, 36, 35, 34, 33, 32, 30, 31}
	for i, f := range feels {
		w.Hourly.Data = append(w.Hourly.Data, weather.DataPoint{
			Time:                float64(start.Add(time.Duration(i) * time.Hour).Unix()),
			Summary:             "Clear",
			Temperature:         f + 3,
			ApparentTemperature: f,
			Humidity:            0.5,
			WindSpeed:           5,
		})
	}
	w.Daily.Data = []weather.DataPoint{
		{Time: float64(start.Unix()), PeriodName: "Tonight", TemperatureMax: 51, TemperatureMin: 37, PrecipProbability: 0.01},
		{Time: float64(at(6, 0).AddDate(0, 0, 1).Unix()), PeriodName: "Wednesday", TemperatureMax: 61, TemperatureMin: 49},
	}
	return w
}

func TestClothingWindowEnd(t *testing.T) {
	tests := []struct {
		now  time.Time
		want time.Time
	}{
		{at(7, 0), at(19, 0)},                   // capped at ClothingHours
		{at(14, 0), at(0, 0).AddDate(0, 0, 1)},  // ends at midnight
		{at(19, 11), at(0, 0).AddDate(0, 0, 1)}, // evening: not tomorrow's pre-dawn
		{at(23, 0), at(2, 0).AddDate(0, 0, 1)},  // late night: at least ClothingMinHours
	}
	for _, tt := range tests {
		if got := clothingWindowEnd(tt.now); !got.Equal(tt.want) {
			t.Errorf("clothingWindowEnd(%s) = %s, want %s", tt.now.Format("15:04"), got.Format("Mon 15:04"), tt.want.Format("Mon 15:04"))
		}
	}
}

// At 19:11 the outfit is for the evening, not for 06:00 tomorrow.
func TestEveningClothingIgnoresPreDawnLow(t *testing.T) {
	withClock(t, at(19, 11))
	rec := GetClothingRecommendation(eveningForecast())
	if rec.Coldest != 42 {
		t.Errorf("coldest feels-like before midnight = %.0f, want 42", rec.Coldest)
	}
	if rec.Outfit != OutfitDescHeavyJacket {
		t.Errorf("outfit = %q, want %q", rec.Outfit, OutfitDescHeavyJacket)
	}
	for _, a := range rec.Accessories {
		if a == "Scarf" {
			t.Errorf("evening at 42F+ should not need a scarf: %v", rec.Accessories)
		}
	}
}

// The layering tip says when the coldest point is.
func TestLayeringTipNamesColdestHour(t *testing.T) {
	withClock(t, at(7, 0))
	w := eveningForecast()
	base := at(7, 0)
	for i := range w.Hourly.Data {
		w.Hourly.Data[i].Time = float64(base.Add(time.Duration(i) * time.Hour).Unix())
	}
	w.Daily.Data[0].PeriodName = "Today"
	rec := GetClothingRecommendation(w)
	want := "feels like 30F at 18:00"
	if len(rec.Notes) == 0 || !strings.Contains(strings.Join(rec.Notes, "|"), want) {
		t.Errorf("notes %q missing %q", rec.Notes, want)
	}
}

// Health warnings outrank comfort tips, since Insights shows only the first note.
func TestAirQualityNoteComesFirst(t *testing.T) {
	withClock(t, at(7, 0))
	w := eveningForecast()
	base := at(7, 0)
	for i := range w.Hourly.Data {
		w.Hourly.Data[i].Time = float64(base.Add(time.Duration(i) * time.Hour).Unix())
		w.Hourly.Data[i].WindGust = 30
	}
	a := []air.Forecast{{DateForecast: "2026-10-06", ParameterName: "PM2.5", AQI: 155, Category: air.Category{Name: "Unhealthy"}}}
	rec := GetClothingRecommendationWithAir(w, a)
	if len(rec.Notes) < 3 || !strings.HasPrefix(rec.Notes[0], "Unhealthy air (PM2.5)") {
		t.Errorf("first note should be the air quality warning, got %q", rec.Notes)
	}
	for _, acc := range rec.Accessories {
		if strings.Contains(strings.ToLower(acc), "mask") {
			t.Errorf("mask advice belongs in notes, not accessories: %q", rec.Accessories)
		}
	}
}

// Feels-like already includes wind chill and heat index; tips must not count them again.
func TestNoDoubleCountedFeelsLikeTips(t *testing.T) {
	w, a := fixtureForecast()
	rec := GetClothingRecommendationWithAir(w, a)
	for _, n := range rec.Notes {
		lower := strings.ToLower(n)
		if strings.Contains(lower, "wind chill") || strings.Contains(lower, "feels warmer") {
			t.Errorf("note double-counts feels-like: %q", n)
		}
	}
}

// After dark, Insights shows tonight's low and tomorrow's high, not a "high" for tonight.
func TestInsightsHighLowAtNight(t *testing.T) {
	withClock(t, at(19, 11))
	out := capture(t, func() { InsightsReport(eveningForecast(), nil) })
	for _, want := range []string{"Low Tonight:", "37F", "High Wednesday:", "61F"} {
		if !strings.Contains(out, want) {
			t.Errorf("insights missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "High / Low") {
		t.Errorf("insights shows High / Low for a night period:\n%s", out)
	}
}

func TestInsightsHighLowByDay(t *testing.T) {
	withClock(t, at(14, 0))
	w := eveningForecast()
	w.Daily.Data[0].PeriodName = "This Afternoon"
	out := capture(t, func() { InsightsReport(w, nil) })
	if !strings.Contains(out, "High / Low:") {
		t.Errorf("daytime insights missing High / Low:\n%s", out)
	}
}

// A night period has no daytime high, so week and daily don't show one for it.
func TestNoHighForNightPeriod(t *testing.T) {
	withClock(t, at(19, 11))
	w := eveningForecast()
	week := capture(t, func() { WeatherWeek(w, nil) })
	if !strings.Contains(week, "Tonight    37   -") {
		t.Errorf("week shows a high for Tonight:\n%s", week)
	}
	daily := capture(t, func() { WeatherDaily(w, nil) })
	tonight, _, _ := strings.Cut(daily, "Wednesday")
	if strings.Contains(tonight, "High:") || !strings.Contains(daily, "High:") {
		t.Errorf("daily should show High for Wednesday only:\n%s", daily)
	}
}

// The clothing range labels the warmest actual and coldest feels-like temperatures separately.
func TestClothingRangeLabels(t *testing.T) {
	withClock(t, at(19, 11))
	out := capture(t, func() { ClothingReport(eveningForecast(), nil) })
	if !strings.Contains(out, "Until 00:00:  up to 54F, feels as cold as 42F at 23:00") {
		t.Errorf("unexpected clothing range:\n%s", out)
	}
}

func TestFormatAlertWindow(t *testing.T) {
	now := at(19, 11)
	tests := []struct {
		name           string
		onset, expires time.Time
		want           string
	}{
		{"not yet started", at(0, 0).AddDate(0, 0, 1), at(5, 0).AddDate(0, 0, 1), "Wed 00:00-05:00 (starts in 4h)"},
		{"in effect", at(18, 0), at(5, 0).AddDate(0, 0, 1), "until Wed 05:00 (in 9h)"},
		{"no onset", time.Time{}, at(22, 0), "until 22:00 (in 2h)"},
		{"no times", time.Time{}, time.Time{}, ""},
	}
	for _, tt := range tests {
		if got := FormatAlertWindow(tt.onset, tt.expires, now); got != tt.want {
			t.Errorf("%s: FormatAlertWindow = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// An alert that hasn't begun counts down to its start, not its end.
func TestAlertsReportBeforeOnset(t *testing.T) {
	withClock(t, at(19, 11))
	w := eveningForecast()
	w.Alerts = []weather.Alert{{
		Title:       "Frost Advisory",
		Description: "* WHAT...Frost. * WHEN...Midnight to 7 AM.",
		Time:        float64(at(0, 0).AddDate(0, 0, 1).Unix()),
		Expires:     float64(at(5, 0).AddDate(0, 0, 1).Unix()),
	}}
	out := capture(t, func() { WeatherAlertsReport(w, nil) })
	if !strings.Contains(out, "Starts In:") || strings.Contains(out, "Time Remaining:") {
		t.Errorf("alert before onset should show Starts In, not Time Remaining:\n%s", out)
	}
	insights := capture(t, func() { InsightsReport(w, nil) })
	if !strings.Contains(insights, "starts in 4h") {
		t.Errorf("insights alert headline should say when it starts:\n%s", insights)
	}
}

func TestBringListReadsAsPhrase(t *testing.T) {
	got := strings.Join(sentenceList([]string{"Winter hat or beanie", "Gloves or mittens", "Scarf"}), ", ")
	if want := "Winter hat or beanie, gloves or mittens, scarf"; got != want {
		t.Errorf("sentenceList = %q, want %q", got, want)
	}
}

// NOAA's isDaytime decides, so a night period with an unusual name still has no high.
func TestNightFlagBeatsPeriodName(t *testing.T) {
	withClock(t, at(19, 11))
	w := eveningForecast()
	w.Daily.Data[0].PeriodName = "Christmas Eve"
	w.Daily.Data[0].Night = true
	if week := capture(t, func() { WeatherWeek(w, nil) }); !strings.Contains(week, "Christmas Eve  37   -") {
		t.Errorf("week shows a high for a night period:\n%s", week)
	}
	if out := capture(t, func() { InsightsReport(w, nil) }); !strings.Contains(out, "Low Christmas Eve:") {
		t.Errorf("insights treats a night period as day:\n%s", out)
	}
}
