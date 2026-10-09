package report

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

// -json is for scripts, so its values match what the text reports show.
func TestJSONMatchesReports(t *testing.T) {
	w, a := fixtureForecast()

	hours := HourlyData(w, a)["hours"].([]HourJSON)
	if len(hours) != DefaultHourlyLimit {
		t.Fatalf("hourly -json has %d hours, want %d", len(hours), DefaultHourlyLimit)
	}
	h := hours[0]
	if h.TempF != 10 || h.FeelsLikeF != -2 || h.PrecipPct != 60 || h.WindDir != "NW" {
		t.Errorf("first hour = %+v", h)
	}
	if h.GustMph == nil || *h.GustMph != 30 || h.DewpointF == nil || *h.DewpointF != -5 {
		t.Errorf("first hour gust/dewpoint = %v/%v, want 30/-5", h.GustMph, h.DewpointF)
	}

	days := DaysData(w, a)["days"].([]DayJSON)
	if len(days) != 7 || days[0].Name != "Today" || days[0].HighF == nil || *days[0].HighF != 12 || days[0].LowF != -5 {
		t.Errorf("days[0] = %+v", days[0])
	}
	if days[0].PressureInHg == nil || *days[0].PressureInHg != 30.02 || days[1].PressureInHg != nil {
		t.Error("pressure should be reported for today only")
	}

	alerts := AlertsData(w, a)["alerts"].([]AlertJSON)
	if len(alerts) != 1 || alerts[0].Event != "Winter Storm Warning" || alerts[0].Onset == nil {
		t.Errorf("alerts = %+v", alerts)
	}

	aqi := InsightsData(w, a)["aqi_today"].(*AQITodayJSON)
	if aqi == nil || aqi.AQI != 155 || aqi.Pollutant != "PM2.5" {
		t.Errorf("aqi_today = %+v, want the day's highest, 155 PM2.5", aqi)
	}

	outfit := ClothingData(w, a)["outfit"].(OutfitJSON)
	hasAirTip := slices.ContainsFunc(outfit.Tips, func(tip string) bool { return strings.HasPrefix(tip, "Unhealthy air") })
	if outfit.Outfit == "" || !hasAirTip {
		t.Errorf("outfit = %+v, want the unhealthy air tip", outfit)
	}
}

// Missing values are null, never a misleading 0.
func TestJSONNulls(t *testing.T) {
	w := eveningForecast()
	tonight := DaysData(w, nil)["days"].([]DayJSON)[0]
	if !tonight.Night || tonight.HighF != nil {
		t.Errorf("Tonight = %+v, want night with a null high", tonight)
	}

	_, a := fixtureForecast()
	rows := AirData(w, a)["air"].([]AirJSON)
	if rows[0].AQI == nil || *rows[0].AQI != 42 || rows[2].AQI != nil {
		t.Errorf("air rows = %+v, want the pending AQI null", rows)
	}

	if got := InsightsData(w, nil)["aqi_today"].(*AQITodayJSON); got != nil {
		t.Errorf("aqi_today with no air data = %+v, want nil", got)
	}

	b, err := json.Marshal(SummaryData(w, nil))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"aqi_today":null`, `"alerts":[]`, `"gust_mph":null`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("summary JSON missing %s:\n%s", want, b)
		}
	}
}
