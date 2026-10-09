package report

import (
	"slices"
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

// noteTexts returns a recommendation's notes as text, in order.
func noteTexts(rec ClothingRecommendation) []string {
	texts := make([]string, len(rec.Notes))
	for i, n := range rec.Notes {
		texts[i] = n.Text
	}
	return texts
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
	rec := GetClothingRecommendation(eveningForecast(), nil)
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
	rec := GetClothingRecommendation(w, nil)
	want := "feels like 30F at 18:00"
	if notes := noteTexts(rec); !strings.Contains(strings.Join(notes, "|"), want) {
		t.Errorf("notes %q missing %q", notes, want)
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
	rec := GetClothingRecommendation(w, a)
	if len(rec.Notes) < 3 || !strings.HasPrefix(rec.Notes[0].Text, "Unhealthy air (PM2.5)") {
		t.Errorf("first note should be the air quality warning, got %q", noteTexts(rec))
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
	rec := GetClothingRecommendation(w, a)
	for _, n := range noteTexts(rec) {
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

// A wind chill of 44.6F shows as 45F, so the outfit and the marked row are 45-54F's.
func TestOutfitMatchesShownTemperature(t *testing.T) {
	withClock(t, at(19, 11))
	w := eveningForecast()
	for i := range w.Hourly.Data {
		w.Hourly.Data[i].ApparentTemperature = max(w.Hourly.Data[i].ApparentTemperature, 44.6)
	}
	if got := GetClothingRecommendation(w, nil).Outfit; got != OutfitDescWarmLayers {
		t.Errorf("outfit = %q, want %q", got, OutfitDescWarmLayers)
	}
	out := capture(t, func() { ClothingReport(w, nil) })
	if !strings.Contains(out, "feels as cold as 45F") || !strings.Contains(out, "> 45-54F") {
		t.Errorf("shown temperature and marked row disagree:\n%s", out)
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

// NOAA's isDaytime decides; period names are the fallback for older caches.
func TestIsNightPeriod(t *testing.T) {
	tests := []struct {
		day  weather.DataPoint
		want bool
	}{
		{weather.DataPoint{PeriodName: "Christmas Eve", Night: true}, true},
		{weather.DataPoint{PeriodName: "Tonight"}, true},
		{weather.DataPoint{PeriodName: "Overnight"}, true},
		{weather.DataPoint{PeriodName: "Monday Night"}, true},
		{weather.DataPoint{PeriodName: "This Afternoon"}, false},
	}
	for _, tt := range tests {
		if got := IsNightPeriod(tt.day); got != tt.want {
			t.Errorf("IsNightPeriod(%q, Night=%v) = %v, want %v", tt.day.PeriodName, tt.day.Night, got, tt.want)
		}
	}
}

// Rain later in the clothing window shows in the Precipitation section, not just in Bring.
func TestClothingPrecipMatchesWindow(t *testing.T) {
	withClock(t, at(14, 0))
	w := eveningForecast()
	base := at(14, 0)
	for i := range w.Hourly.Data {
		w.Hourly.Data[i].Time = float64(base.Add(time.Duration(i) * time.Hour).Unix())
		if i >= 7 && i <= 9 {
			w.Hourly.Data[i].PrecipProbability, w.Hourly.Data[i].PrecipType = 0.8, "rain"
		}
	}
	w.Daily.Data[0].PeriodName = "This Afternoon"
	w.Daily.Data[0].PrecipProbability = 0
	out := capture(t, func() { ClothingReport(w, nil) })
	if !strings.Contains(out, "Umbrella") || !strings.Contains(out, "80% chance of rain") || strings.Contains(out, "No precipitation expected") {
		t.Errorf("umbrella advice without the rain that explains it:\n%s", out)
	}
}

// A precipitation chance shows once, in its own section, whatever NOAA calls the type.
func TestClothingPrecipTipNotRepeated(t *testing.T) {
	withClock(t, at(14, 0))
	w := eveningForecast()
	base := at(14, 0)
	for i := range w.Hourly.Data {
		w.Hourly.Data[i].Time = float64(base.Add(time.Duration(i) * time.Hour).Unix())
		w.Hourly.Data[i].PrecipProbability, w.Hourly.Data[i].PrecipType = 0.75, "mix"
	}
	w.Daily.Data[0].PeriodName = "This Afternoon"
	out := capture(t, func() { ClothingReport(w, nil) })
	if n := strings.Count(out, "75% chance of mix"); n != 1 {
		t.Errorf("precipitation chance appears %d times, want once:\n%s", n, out)
	}
}

func TestHoursMinutes(t *testing.T) {
	for d, want := range map[time.Duration]string{
		time.Minute:                  "1 minute",
		48 * time.Minute:             "48 minutes",
		time.Hour:                    "1 hour, 0 minutes",
		2*time.Hour + time.Minute:    "2 hours, 1 minute",
		4*time.Hour + 49*time.Minute: "4 hours, 49 minutes",
	} {
		if got := hoursMinutes(d); got != want {
			t.Errorf("hoursMinutes(%v) = %q, want %q", d, got, want)
		}
	}
}

// A saved forecast shown later drops what has passed, and leaves the saved slices alone.
func TestDropPast(t *testing.T) {
	now := at(14, 0)
	hour := func(h int) float64 { return float64(now.Add(time.Duration(h) * time.Hour).Unix()) }
	var w weather.Forecast
	w.Daily.Data = []weather.DataPoint{
		{Time: hour(-20), PeriodName: "Tonight"}, // last night, over once Today began
		{Time: hour(-8), PeriodName: "Today"},
		{Time: hour(4), PeriodName: "Tonight"},
	}
	w.Alerts = []weather.Alert{
		{Title: "Frost Advisory", Expires: hour(-9)},
		{Title: "Wind Advisory", Expires: hour(2)},
		{Title: "Flood Watch"}, // no expiry given
	}
	a := []air.Forecast{
		{DateForecast: "2026-10-05 ", ParameterName: "OZONE"},
		{DateForecast: "2026-10-06 ", ParameterName: "OZONE"},
		{DateForecast: "2026-10-07 ", ParameterName: "OZONE"},
	}
	daily, alerts, airRows := slices.Clone(w.Daily.Data), slices.Clone(w.Alerts), slices.Clone(a)

	gotW, gotA := DropPast(w, a, now)
	if len(gotW.Daily.Data) != 2 || gotW.Daily.Data[0].PeriodName != "Today" {
		t.Errorf("daily = %+v, want Today first", gotW.Daily.Data)
	}
	if len(gotW.Alerts) != 2 || gotW.Alerts[0].Title != "Wind Advisory" {
		t.Errorf("alerts = %+v, want the wind advisory and flood watch", gotW.Alerts)
	}
	if len(gotA) != 2 || strings.TrimSpace(gotA[0].DateForecast) != "2026-10-06" {
		t.Errorf("air = %+v, want today onward", gotA)
	}
	if !slices.Equal(w.Daily.Data, daily) || !slices.Equal(w.Alerts, alerts) || !slices.Equal(a, airRows) {
		t.Error("DropPast modified its input slices, which main saves to the cache")
	}

	// Every period over: the last is kept, since reports read Daily.Data[0].
	w.Daily.Data = w.Daily.Data[:2]
	if gotW, _ := DropPast(w, nil, now.Add(24*time.Hour)); len(gotW.Daily.Data) != 1 {
		t.Errorf("all periods past: got %d, want the last one kept", len(gotW.Daily.Data))
	}
}

// AirNow dates are the location's: 01:00 in New York is still yesterday in Los Angeles.
func TestDropPastUsesForecastTimezone(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("no tz database:", err)
	}
	now := time.Date(2026, time.October, 7, 1, 0, 0, 0, ny)
	w := weather.Forecast{Timezone: "America/Los_Angeles"}
	a := []air.Forecast{{DateForecast: "2026-10-06", ParameterName: "OZONE"}}
	if _, got := DropPast(w, a, now); len(got) != 1 {
		t.Errorf("dropped Los Angeles's current day: %+v", got)
	}
}
