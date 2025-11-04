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

// TestSafeSliceHourly tests safe array slicing with bounds checking
func TestSafeSliceHourly(t *testing.T) {
	tests := []struct {
		name     string
		data     []weather.DataPoint
		maxHours int
		wantLen  int
	}{
		{
			name:     "requested 12, have 20",
			data:     make([]weather.DataPoint, 20),
			maxHours: 12,
			wantLen:  12,
		},
		{
			name:     "requested 12, have 8",
			data:     make([]weather.DataPoint, 8),
			maxHours: 12,
			wantLen:  8,
		},
		{
			name:     "requested 12, have 0",
			data:     []weather.DataPoint{},
			maxHours: 12,
			wantLen:  0,
		},
		{
			name:     "empty slice",
			data:     nil,
			maxHours: 12,
			wantLen:  0,
		},
		{
			name:     "requested 0",
			data:     make([]weather.DataPoint, 10),
			maxHours: 0,
			wantLen:  0,
		},
		{
			name:     "requested negative",
			data:     make([]weather.DataPoint, 10),
			maxHours: -5,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeSliceHourly(tt.data, tt.maxHours)
			if len(got) != tt.wantLen {
				t.Errorf("SafeSliceHourly() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// TestIsPrecipitationNote tests precipitation keyword detection
func TestIsPrecipitationNote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"umbrella mention", "Bring an umbrella", true},
		{"precipitation keyword", "Watch for precipitation", true},
		{"rain keyword", "Rain expected later", true},
		{"snow keyword", "Snow possible overnight", true},
		{"sleet keyword", "Sleet and ice", true},
		{"sunny weather", "It's sunny", false},
		{"empty string", "", false},
		{"wind only", "Strong winds expected", false},
		{"mixed case rain", "RAIN warning", true},
		{"mixed case umbrella", "UMBRELLA recommended", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPrecipitationNote(tt.input)
			if got != tt.want {
				t.Errorf("IsPrecipitationNote(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
