package report

import (
	"testing"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// TestGetCurrentHourlyData tests finding the most current hourly data point
func TestGetCurrentHourlyData(t *testing.T) {
	tests := []struct {
		name     string
		forecast weather.Forecast
		wantFrom string // "currently" or "hourly"
	}{
		{
			name: "empty hourly data returns Currently",
			forecast: weather.Forecast{
				Currently: weather.DataPoint{Temperature: 75.0},
				Hourly:    weather.DataBlock{Data: []weather.DataPoint{}},
			},
			wantFrom: "currently",
		},
		{
			name: "past hours only returns Currently",
			forecast: weather.Forecast{
				Currently: weather.DataPoint{Temperature: 75.0},
				Hourly: weather.DataBlock{
					Data: []weather.DataPoint{
						{Time: float64(time.Now().Unix()) - 3600, Temperature: 70.0}, // 1 hour ago
						{Time: float64(time.Now().Unix()) - 7200, Temperature: 68.0}, // 2 hours ago
					},
				},
			},
			wantFrom: "currently",
		},
		{
			name: "future hours returns nearest future hour",
			forecast: weather.Forecast{
				Currently: weather.DataPoint{Temperature: 75.0},
				Hourly: weather.DataBlock{
					Data: []weather.DataPoint{
						{Time: float64(time.Now().Unix()) + 1800, Temperature: 76.0}, // 30 min future
						{Time: float64(time.Now().Unix()) + 3600, Temperature: 77.0}, // 1 hour future
						{Time: float64(time.Now().Unix()) + 7200, Temperature: 78.0}, // 2 hours future
					},
				},
			},
			wantFrom: "hourly",
		},
		{
			name: "mixed past and future returns nearest future",
			forecast: weather.Forecast{
				Currently: weather.DataPoint{Temperature: 75.0},
				Hourly: weather.DataBlock{
					Data: []weather.DataPoint{
						{Time: float64(time.Now().Unix()) - 3600, Temperature: 70.0}, // past
						{Time: float64(time.Now().Unix()) + 1800, Temperature: 76.0}, // future - nearest
						{Time: float64(time.Now().Unix()) + 3600, Temperature: 77.0}, // future
					},
				},
			},
			wantFrom: "hourly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCurrentHourlyData(tt.forecast)
			if tt.wantFrom == "currently" {
				if got.Temperature != tt.forecast.Currently.Temperature {
					t.Errorf("GetCurrentHourlyData() = %v, want Currently with temp %v",
						got.Temperature, tt.forecast.Currently.Temperature)
				}
			} else {
				// Should get first future hourly value
				if len(tt.forecast.Hourly.Data) > 0 {
					// Find first future entry
					currentTime := float64(time.Now().Unix())
					var expectedTemp float64
					for _, hour := range tt.forecast.Hourly.Data {
						if hour.Time >= currentTime {
							expectedTemp = hour.Temperature
							break
						}
					}
					if got.Temperature != expectedTemp {
						t.Errorf("GetCurrentHourlyData() temp = %v, want %v from nearest future hour",
							got.Temperature, expectedTemp)
					}
				}
			}
		})
	}
}

// TestCapitalizeFirst tests string capitalization
func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"single lowercase", "a", "A"},
		{"single uppercase", "A", "A"},
		{"rain/snow special case", "rain/snow", "Rain/Snow"},
		{"rain", "rain", "Rain"},
		{"snow", "snow", "Snow"},
		{"already uppercase", "RAIN", "RAIN"},
		{"sleet", "sleet", "Sleet"},
		{"multi-word", "freezing rain", "Freezing rain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CapitalizeFirst(tt.input)
			if got != tt.want {
				t.Errorf("CapitalizeFirst(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestGetPrecipTypeOrDefault tests precipitation type formatting
func TestGetPrecipTypeOrDefault(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", "Precipitation"},
		{"rain", "rain", "Rain"},
		{"snow", "snow", "Snow"},
		{"rain/snow", "rain/snow", "Rain/Snow"},
		{"sleet", "sleet", "Sleet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPrecipTypeOrDefault(tt.input)
			if got != tt.want {
				t.Errorf("GetPrecipTypeOrDefault(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseNWSDescription tests splitting NWS alert text into labeled sections
func TestParseNWSDescription(t *testing.T) {
	desc := "* WHAT...Temperatures as low as 33 will result in frost\nformation.\n\n* WHERE...The Champlain Valley.\n\n* WHEN...From midnight tonight to 7 AM EDT Wednesday.\n\n* ADDITIONAL DETAILS...Cover plants."
	got := ParseNWSDescription(desc)
	want := []LabeledText{
		{"What", "Temperatures as low as 33 will result in frost formation."},
		{"Where", "The Champlain Valley."},
		{"When", "From midnight tonight to 7 AM EDT Wednesday."},
		{"Additional details", "Cover plants."},
	}
	if len(got) != len(want) {
		t.Fatalf("ParseNWSDescription() returned %d sections, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("section %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	for _, unstructured := range []string{"", "A plain description.", "* not uppercase...text"} {
		if got := ParseNWSDescription(unstructured); got != nil {
			t.Errorf("ParseNWSDescription(%q) = %+v, want nil", unstructured, got)
		}
	}
}

// TestWrapTextCountsRunes ensures multi-byte characters count as one column
func TestWrapTextCountsRunes(t *testing.T) {
	// 10 runes but 11 bytes because of the degree sign
	if got := WrapText("High 55°F.", 10); len(got) != 1 {
		t.Errorf("WrapText() = %q, want a single line", got)
	}
}

// TestFormatUntil tests clock time with relative offsets
func TestFormatUntil(t *testing.T) {
	now := time.Date(2026, 10, 6, 14, 0, 0, 0, time.Local)
	tests := []struct {
		t    time.Time
		want string
	}{
		{now.Add(30 * time.Minute), "14:30 (in 30m)"},
		{now.Add(5 * time.Hour), "19:00 (in 5h)"},
		{now.Add(15 * time.Hour), "Wed 05:00 (in 15h)"},
		{now.Add(72 * time.Hour), "Fri 14:00 (in 3d)"},
		{now.Add(-time.Hour), "13:00"},
	}
	for _, tt := range tests {
		if got := FormatUntil(tt.t, now); got != tt.want {
			t.Errorf("FormatUntil(%v) = %q, want %q", tt.t, got, tt.want)
		}
	}
}
