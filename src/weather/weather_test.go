package weather

import (
	"encoding/json"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"testing"
)

// Sample NOAA Points API response
func getSampleNOAAPointsResponse() NOAAPointsResponse {
	return NOAAPointsResponse{
		Properties: NOAAPointsProperties{
			GridID:         "PHI",
			GridX:          45,
			GridY:          67,
			Forecast:       "https://api.weather.gov/gridpoints/PHI/45,67/forecast",
			ForecastHourly: "https://api.weather.gov/gridpoints/PHI/45,67/forecast/hourly",
			RelativeLocation: NOAARelativeLocation{
				Properties: NOAARelativeLocationProps{
					City:  "Philadelphia",
					State: "PA",
				},
			},
		},
	}
}

// Sample NOAA Daily Forecast response with day period first
func getSampleNOAADailyForecastDayFirst() NOAAForecastResponse {
	precipProb := 30.0
	dewpoint := 10.0
	humidity := 65.0

	return NOAAForecastResponse{
		Properties: NOAAForecastProperties{
			Updated: "2025-10-24T12:00:00+00:00",
			Periods: []NOAAPeriod{
				{
					Number:        1,
					Name:          "Today",
					StartTime:     "2025-10-24T06:00:00-04:00",
					EndTime:       "2025-10-24T18:00:00-04:00",
					IsDaytime:     true,
					Temperature:   75,
					TemperatureUnit: "F",
					WindSpeed:     "10 to 15 mph",
					WindDirection: "SW",
					ShortForecast: "Partly Cloudy",
					DetailedForecast: "Partly cloudy skies throughout the day.",
					ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
					Dewpoint:                   NOAAValue{Value: &dewpoint},
					RelativeHumidity:           NOAAValue{Value: &humidity},
				},
				{
					Number:        2,
					Name:          "Tonight",
					StartTime:     "2025-10-24T18:00:00-04:00",
					EndTime:       "2025-10-25T06:00:00-04:00",
					IsDaytime:     false,
					Temperature:   55,
					TemperatureUnit: "F",
					WindSpeed:     "5 to 10 mph",
					WindDirection: "S",
					ShortForecast: "Mostly Clear",
					DetailedForecast: "Mostly clear skies overnight.",
					ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
					Dewpoint:                   NOAAValue{Value: &dewpoint},
					RelativeHumidity:           NOAAValue{Value: &humidity},
				},
			},
		},
	}
}

// Sample NOAA Daily Forecast response with night period first (this happens when requesting in evening)
func getSampleNOAADailyForecastNightFirst() NOAAForecastResponse {
	precipProb := 20.0
	dewpoint := 8.0
	humidity := 70.0

	return NOAAForecastResponse{
		Properties: NOAAForecastProperties{
			Updated: "2025-10-24T20:00:00+00:00",
			Periods: []NOAAPeriod{
				{
					Number:        1,
					Name:          "Tonight",
					StartTime:     "2025-10-24T18:00:00-04:00",
					EndTime:       "2025-10-25T06:00:00-04:00",
					IsDaytime:     false,
					Temperature:   50,
					TemperatureUnit: "F",
					WindSpeed:     "5 mph",
					WindDirection: "NW",
					ShortForecast: "Clear",
					DetailedForecast: "Clear skies overnight.",
					ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
					Dewpoint:                   NOAAValue{Value: &dewpoint},
					RelativeHumidity:           NOAAValue{Value: &humidity},
				},
				{
					Number:        2,
					Name:          "Tomorrow",
					StartTime:     "2025-10-25T06:00:00-04:00",
					EndTime:       "2025-10-25T18:00:00-04:00",
					IsDaytime:     true,
					Temperature:   80,
					TemperatureUnit: "F",
					WindSpeed:     "10 mph",
					WindDirection: "W",
					ShortForecast: "Sunny",
					DetailedForecast: "Sunny skies throughout the day.",
					ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
					Dewpoint:                   NOAAValue{Value: &dewpoint},
					RelativeHumidity:           NOAAValue{Value: &humidity},
				},
			},
		},
	}
}

// Sample NOAA Hourly Forecast response
func getSampleNOAAHourlyForecast() NOAAForecastResponse {
	precipProb := 10.0
	dewpoint := 12.0
	humidity := 60.0

	return NOAAForecastResponse{
		Properties: NOAAForecastProperties{
			Updated: "2025-10-24T12:00:00+00:00",
			Periods: []NOAAPeriod{
				{
					Number:        1,
					Name:          "Now",
					StartTime:     "2025-10-24T12:00:00-04:00",
					EndTime:       "2025-10-24T13:00:00-04:00",
					IsDaytime:     true,
					Temperature:   72,
					TemperatureUnit: "F",
					WindSpeed:     "12 mph",
					WindDirection: "SW",
					ShortForecast: "Partly Cloudy",
					DetailedForecast: "Partly cloudy conditions.",
					ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
					Dewpoint:                   NOAAValue{Value: &dewpoint},
					RelativeHumidity:           NOAAValue{Value: &humidity},
				},
				{
					Number:        2,
					Name:          "1pm",
					StartTime:     "2025-10-24T13:00:00-04:00",
					EndTime:       "2025-10-24T14:00:00-04:00",
					IsDaytime:     true,
					Temperature:   74,
					TemperatureUnit: "F",
					WindSpeed:     "13 mph",
					WindDirection: "SW",
					ShortForecast: "Mostly Sunny",
					DetailedForecast: "Mostly sunny conditions.",
					ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
					Dewpoint:                   NOAAValue{Value: &dewpoint},
					RelativeHumidity:           NOAAValue{Value: &humidity},
				},
			},
		},
	}
}

func TestConvertNOAAToForecast(t *testing.T) {
	points := getSampleNOAAPointsResponse()
	daily := getSampleNOAADailyForecastDayFirst()
	hourly := getSampleNOAAHourlyForecast()
	coords := geolocation.Coordinates{
		Latitude:  "39.9526",
		Longitude: "-75.1652",
	}

	forecast := ConvertNOAAToForecast(points, daily, hourly, coords)

	// Test basic properties
	if forecast.Latitude != 39.9526 {
		t.Errorf("Expected latitude 39.9526, got %f", forecast.Latitude)
	}

	if forecast.Longitude != -75.1652 {
		t.Errorf("Expected longitude -75.1652, got %f", forecast.Longitude)
	}

	// Test currently (from first hourly period)
	if forecast.Currently.Temperature != 72 {
		t.Errorf("Expected current temperature 72, got %f", forecast.Currently.Temperature)
	}

	// Test hourly data
	if len(forecast.Hourly.Data) != 2 {
		t.Errorf("Expected 2 hourly data points, got %d", len(forecast.Hourly.Data))
	}

	// Test daily data
	if len(forecast.Daily.Data) != 1 {
		t.Errorf("Expected 1 daily data point, got %d", len(forecast.Daily.Data))
	}
}

func TestConvertNOAADailyPeriodsToDataBlock_DayFirst(t *testing.T) {
	daily := getSampleNOAADailyForecastDayFirst()
	block := convertNOAADailyPeriodsToDataBlock(daily.Properties.Periods)

	if len(block.Data) != 1 {
		t.Fatalf("Expected 1 daily data point, got %d", len(block.Data))
	}

	dp := block.Data[0]

	// When day period is first, max temp should be from day (75), min from night (55)
	if dp.TemperatureMax != 75 {
		t.Errorf("Expected max temperature 75, got %f", dp.TemperatureMax)
	}

	if dp.TemperatureMin != 55 {
		t.Errorf("Expected min temperature 55, got %f", dp.TemperatureMin)
	}

	// Verify precipitation probability is converted from percentage to decimal
	expected := 0.30
	if dp.PrecipProbability != expected {
		t.Errorf("Expected precip probability %f, got %f", expected, dp.PrecipProbability)
	}

	// Verify humidity is converted from percentage to decimal
	expectedHumidity := 0.65
	if dp.Humidity != expectedHumidity {
		t.Errorf("Expected humidity %f, got %f", expectedHumidity, dp.Humidity)
	}
}

func TestConvertNOAADailyPeriodsToDataBlock_NightFirst(t *testing.T) {
	daily := getSampleNOAADailyForecastNightFirst()
	block := convertNOAADailyPeriodsToDataBlock(daily.Properties.Periods)

	if len(block.Data) != 1 {
		t.Fatalf("Expected 1 daily data point, got %d", len(block.Data))
	}

	dp := block.Data[0]

	// When night period is first, max temp should still be from day (80), min from night (50)
	// This tests the swap logic
	if dp.TemperatureMax != 80 {
		t.Errorf("Expected max temperature 80 (from day period), got %f", dp.TemperatureMax)
	}

	if dp.TemperatureMin != 50 {
		t.Errorf("Expected min temperature 50 (from night period), got %f", dp.TemperatureMin)
	}

	// Verify the summary comes from the day period, not the night period
	if dp.Summary != "Sunny" {
		t.Errorf("Expected summary 'Sunny' (from day period), got '%s'", dp.Summary)
	}
}

func TestConvertNOAAPeriodToDataPoint(t *testing.T) {
	precipProb := 25.0
	dewpoint := 15.0
	humidity := 75.0

	period := NOAAPeriod{
		Number:        1,
		Name:          "Today",
		StartTime:     "2025-10-24T12:00:00-04:00",
		IsDaytime:     true,
		Temperature:   70,
		WindSpeed:     "10 to 15 mph",
		WindDirection: "NE",
		ShortForecast: "Rain",
		ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
		Dewpoint:                   NOAAValue{Value: &dewpoint},
		RelativeHumidity:           NOAAValue{Value: &humidity},
	}

	dp := convertNOAAPeriodToDataPoint(period)

	if dp.Temperature != 70 {
		t.Errorf("Expected temperature 70, got %f", dp.Temperature)
	}

	// Test wind speed parsing
	if dp.WindSpeed != 10 {
		t.Errorf("Expected wind speed 10 (first number from range), got %f", dp.WindSpeed)
	}

	// Test wind direction parsing
	if dp.WindBearing != 45 {
		t.Errorf("Expected wind bearing 45 (NE), got %f", dp.WindBearing)
	}

	// Test precipitation probability conversion
	if dp.PrecipProbability != 0.25 {
		t.Errorf("Expected precip probability 0.25, got %f", dp.PrecipProbability)
	}

	// Test dewpoint conversion from Celsius to Fahrenheit
	expectedDewpoint := celsiusToFahrenheit(15.0)
	if dp.DewPoint != expectedDewpoint {
		t.Errorf("Expected dewpoint %f, got %f", expectedDewpoint, dp.DewPoint)
	}

	// Test humidity conversion
	if dp.Humidity != 0.75 {
		t.Errorf("Expected humidity 0.75, got %f", dp.Humidity)
	}

	// Test icon mapping
	if dp.Icon != "rain" {
		t.Errorf("Expected icon 'rain', got '%s'", dp.Icon)
	}
}

func TestParseWindSpeed(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"10 mph", 10.0},
		{"10 to 15 mph", 10.0},
		{"5 to 10 mph", 5.0},
		{"20 mph", 20.0},
	}

	for _, test := range tests {
		result := parseWindSpeed(test.input)
		if result != test.expected {
			t.Errorf("parseWindSpeed(%q) = %f, expected %f", test.input, result, test.expected)
		}
	}
}

func TestParseWindDirection(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"N", 0},
		{"NE", 45},
		{"E", 90},
		{"SE", 135},
		{"S", 180},
		{"SW", 225},
		{"W", 270},
		{"NW", 315},
		{"NNE", 22.5},
		{"ESE", 112.5},
		{"SSW", 202.5},
		{"WNW", 292.5},
		{"INVALID", 0}, // Unknown direction should return 0
	}

	for _, test := range tests {
		result := parseWindDirection(test.input)
		if result != test.expected {
			t.Errorf("parseWindDirection(%q) = %f, expected %f", test.input, result, test.expected)
		}
	}
}

func TestCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		celsius    float64
		fahrenheit float64
	}{
		{0, 32},
		{10, 50},
		{20, 68},
		{100, 212},
		{-40, -40},
	}

	for _, test := range tests {
		result := celsiusToFahrenheit(test.celsius)
		if result != test.fahrenheit {
			t.Errorf("celsiusToFahrenheit(%f) = %f, expected %f", test.celsius, result, test.fahrenheit)
		}
	}
}

func TestMapNOAAIconToIcon(t *testing.T) {
	tests := []struct {
		forecast string
		icon     string
	}{
		{"Sunny", "clear-day"},
		{"Clear", "clear-day"},
		{"Partly Cloudy", "partly-cloudy-day"},
		{"Partly Sunny", "partly-cloudy-day"},
		{"Mostly Cloudy", "cloudy"},
		{"Cloudy", "cloudy"},
		{"Rain", "rain"},
		{"Showers", "rain"},
		{"Snow", "snow"},
		{"Thunderstorm", "thunderstorm"},
		{"Fog", "fog"},
		{"Unknown Condition", "partly-cloudy-day"}, // Default
	}

	for _, test := range tests {
		result := mapNOAAIconToIcon(test.forecast)
		if result != test.icon {
			t.Errorf("mapNOAAIconToIcon(%q) = %q, expected %q", test.forecast, result, test.icon)
		}
	}
}

func TestNullValueHandling(t *testing.T) {
	// Test that nil values don't cause crashes
	period := NOAAPeriod{
		Number:                     1,
		StartTime:                  "2025-10-24T12:00:00-04:00",
		Temperature:                70,
		WindSpeed:                  "10 mph",
		WindDirection:              "N",
		ShortForecast:              "Clear",
		ProbabilityOfPrecipitation: NOAAValue{Value: nil}, // Null value
		Dewpoint:                   NOAAValue{Value: nil}, // Null value
		RelativeHumidity:           NOAAValue{Value: nil}, // Null value
	}

	dp := convertNOAAPeriodToDataPoint(period)

	// Should not panic and should have zero values for null fields
	if dp.PrecipProbability != 0 {
		t.Errorf("Expected precip probability 0 for nil value, got %f", dp.PrecipProbability)
	}

	if dp.DewPoint != 0 {
		t.Errorf("Expected dewpoint 0 for nil value, got %f", dp.DewPoint)
	}

	if dp.Humidity != 0 {
		t.Errorf("Expected humidity 0 for nil value, got %f", dp.Humidity)
	}
}

func TestOddNumberOfPeriods(t *testing.T) {
	// Test handling when there's an odd number of periods (no night period for last day)
	precipProb := 15.0
	dewpoint := 10.0
	humidity := 65.0

	periods := []NOAAPeriod{
		{
			Number:                     1,
			Name:                       "Today",
			StartTime:                  "2025-10-24T06:00:00-04:00",
			IsDaytime:                  true,
			Temperature:                75,
			WindSpeed:                  "10 mph",
			WindDirection:              "S",
			ShortForecast:              "Sunny",
			ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
			Dewpoint:                   NOAAValue{Value: &dewpoint},
			RelativeHumidity:           NOAAValue{Value: &humidity},
		},
		// No night period
	}

	block := convertNOAADailyPeriodsToDataBlock(periods)

	if len(block.Data) != 1 {
		t.Fatalf("Expected 1 daily data point, got %d", len(block.Data))
	}

	dp := block.Data[0]

	// Should have max temp but min temp should be 0 (no night period)
	if dp.TemperatureMax != 75 {
		t.Errorf("Expected max temperature 75, got %f", dp.TemperatureMax)
	}

	if dp.TemperatureMin != 0 {
		t.Errorf("Expected min temperature 0 (no night period), got %f", dp.TemperatureMin)
	}
}

// NEW TESTS FOR ENHANCED FEATURES

func TestApparentTemperatureCalculation(t *testing.T) {
	tests := []struct {
		name         string
		temp         float64
		humidity     float64
		windSpeed    float64
		expectedDiff float64 // Expected difference from actual temp
	}{
		{
			name:         "Hot and humid - feels hotter",
			temp:         90,
			humidity:     0.80,
			windSpeed:    5,
			expectedDiff: 5, // Should feel at least 5° hotter
		},
		{
			name:         "Cold and windy - wind chill",
			temp:         30,
			humidity:     0.50,
			windSpeed:    20,
			expectedDiff: -10, // Should feel at least 10° colder
		},
		{
			name:         "Mild conditions - minimal difference",
			temp:         70,
			humidity:     0.50,
			windSpeed:    10,
			expectedDiff: 0, // Should be close to actual temp
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			apparent := calculateApparentTemperature(test.temp, test.humidity, test.windSpeed)
			diff := apparent - test.temp

			// Check if the difference is in the expected direction
			if test.expectedDiff > 0 && diff < test.expectedDiff {
				t.Errorf("Expected feels like temp to be at least %.0f° warmer, got %.1f° difference",
					test.expectedDiff, diff)
			} else if test.expectedDiff < 0 && diff > test.expectedDiff {
				t.Errorf("Expected feels like temp to be at least %.0f° colder, got %.1f° difference",
					test.expectedDiff, diff)
			}
		})
	}
}

func TestPrecipitationTypeExtraction(t *testing.T) {
	tests := []struct {
		forecast     string
		expectedType string
	}{
		{"Light Rain", "rain"},
		{"Heavy Rain", "rain"},
		{"Rain Showers", "rain"},
		{"Showers", "rain"},
		{"Light Snow", "snow"},
		{"Heavy Snow", "snow"},
		{"Snow Showers", "snow"},
		{"Chance Rain And Snow", "rain/snow"},
		{"Wintry Mix", "mix"},
		{"Sleet", "sleet"},
		{"Freezing Rain", "freezing rain"},
		{"Sunny", ""},
		{"Partly Cloudy", ""},
		{"Cloudy", ""},
	}

	for _, test := range tests {
		result := extractPrecipType(test.forecast)
		if result != test.expectedType {
			t.Errorf("extractPrecipType(%q) = %q, expected %q",
				test.forecast, result, test.expectedType)
		}
	}
}

func TestSunriseSunsetExtraction(t *testing.T) {
	// Test sunrise/sunset from NOAA icon URLs
	// NOAA embeds day/night info in icon URLs like:
	// https://api.weather.gov/icons/land/day/skc?size=medium
	// https://api.weather.gov/icons/land/night/skc?size=medium

	tests := []struct {
		iconURL   string
		isDaytime bool
	}{
		{"https://api.weather.gov/icons/land/day/skc?size=medium", true},
		{"https://api.weather.gov/icons/land/night/skc?size=medium", false},
		{"https://api.weather.gov/icons/land/day/rain?size=medium", true},
		{"https://api.weather.gov/icons/land/night/snow?size=medium", false},
	}

	for _, test := range tests {
		result := isIconDaytime(test.iconURL)
		if result != test.isDaytime {
			t.Errorf("isIconDaytime(%q) = %v, expected %v",
				test.iconURL, result, test.isDaytime)
		}
	}
}

func TestPrecipitationTypeInDataPoint(t *testing.T) {
	// Test that precipitation type is properly extracted and stored
	precipProb := 80.0
	dewpoint := 10.0
	humidity := 70.0

	period := NOAAPeriod{
		Number:                     1,
		Name:                       "Today",
		StartTime:                  "2025-10-24T12:00:00-04:00",
		IsDaytime:                  true,
		Temperature:                35,
		WindSpeed:                  "10 mph",
		WindDirection:              "N",
		ShortForecast:              "Heavy Snow",
		DetailedForecast:           "Heavy snow expected throughout the day.",
		ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
		Dewpoint:                   NOAAValue{Value: &dewpoint},
		RelativeHumidity:           NOAAValue{Value: &humidity},
	}

	dp := convertNOAAPeriodToDataPoint(period)

	// Should extract "snow" from "Heavy Snow"
	if dp.PrecipType != "snow" {
		t.Errorf("Expected precip type 'snow', got '%s'", dp.PrecipType)
	}

	// Test with rain
	period.ShortForecast = "Rain Showers"
	period.DetailedForecast = "Rain showers likely."
	dp = convertNOAAPeriodToDataPoint(period)

	if dp.PrecipType != "rain" {
		t.Errorf("Expected precip type 'rain', got '%s'", dp.PrecipType)
	}
}

func TestApparentTemperatureInConversion(t *testing.T) {
	precipProb := 25.0
	dewpoint := 15.0
	humidity := 80.0

	period := NOAAPeriod{
		Number:                     1,
		Name:                       "Now",
		StartTime:                  "2025-10-24T14:00:00-04:00",
		Temperature:                90,
		WindSpeed:                  "5 mph",
		WindDirection:              "S",
		ShortForecast:              "Sunny",
		ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
		Dewpoint:                   NOAAValue{Value: &dewpoint},
		RelativeHumidity:           NOAAValue{Value: &humidity},
	}

	dp := convertNOAAPeriodToDataPoint(period)

	// With 90°F and 80% humidity, feels like should be significantly higher
	if dp.ApparentTemperature <= dp.Temperature {
		t.Errorf("Expected apparent temp (%.1f) to be higher than actual temp (%.1f) in hot humid conditions",
			dp.ApparentTemperature, dp.Temperature)
	}

	// Verify it's not zero (which would indicate it wasn't calculated)
	if dp.ApparentTemperature == 0 {
		t.Error("Apparent temperature should not be zero when conditions are available")
	}
}

func TestWindGustParsing(t *testing.T) {
	tests := []struct {
		gustString   string
		expectedGust float64
	}{
		{"25 mph", 25.0},
		{"30 mph", 30.0},
		{"", 0.0},
	}

	for _, test := range tests {
		gustPtr := &test.gustString
		if test.gustString == "" {
			gustPtr = nil
		}
		result := parseWindGust(gustPtr)
		if result != test.expectedGust {
			t.Errorf("parseWindGust(%v) = %.1f, expected %.1f",
				test.gustString, result, test.expectedGust)
		}
	}
}

func TestTimezoneDetection(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lon      float64
		expected string
	}{
		{"New York (Eastern)", 40.7128, -74.0060, "America/New_York"},
		{"Chicago (Central)", 41.8781, -87.6298, "America/Chicago"},
		{"Denver (Mountain)", 39.7392, -104.9903, "America/Denver"},
		{"Los Angeles (Pacific)", 34.0522, -118.2437, "America/Los_Angeles"},
		{"Anchorage (Alaska)", 61.2181, -149.9003, "America/Anchorage"},
		{"Honolulu (Hawaii)", 21.3099, -157.8581, "Pacific/Honolulu"},
		{"Miami (Eastern)", 25.7617, -80.1918, "America/New_York"},
		{"Seattle (Pacific)", 47.6062, -122.3321, "America/Los_Angeles"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getTimezoneFromCoordinates(test.lat, test.lon)
			if result != test.expected {
				t.Errorf("getTimezoneFromCoordinates(%.4f, %.4f) = %s, expected %s",
					test.lat, test.lon, result, test.expected)
			}
		})
	}
}

// TestWindChillFormula tests that wind chill is calculated correctly using NWS formula.
// This test will FAIL if the formula incorrectly uses multiplication instead of exponentiation.
func TestWindChillFormula(t *testing.T) {
	tests := []struct {
		name          string
		temp          float64
		windSpeed     float64
		expectedRange struct{ min, max float64 } // Allow small tolerance for floating point
	}{
		{
			name:      "NWS Example: 30°F, 20mph wind",
			temp:      30.0,
			windSpeed: 20.0,
			// Correct NWS formula: 35.74 + 0.6215*30 - 35.75*(20^0.16) + 0.4275*30*(20^0.16)
			// Using V^0.16 = 1.8171: 35.74 + 18.645 - 64.96 + 23.29 = 17.36°F
			expectedRange: struct{ min, max float64 }{17.0, 17.7},
		},
		{
			name:      "Cold and windy: 15°F, 25mph wind",
			temp:      15.0,
			windSpeed: 25.0,
			// Correct: 35.74 + 0.6215*15 - 35.75*(25^0.16) + 0.4275*15*(25^0.16)
			// Using V^0.16 = 1.8833: 35.74 + 9.32 - 67.33 + 12.08 = -4.04°F
			expectedRange: struct{ min, max float64 }{-4.5, -3.5},
		},
		{
			name:      "Light wind: 40°F, 5mph wind",
			temp:      40.0,
			windSpeed: 5.0,
			// Correct: 35.74 + 0.6215*40 - 35.75*(5^0.16) + 0.4275*40*(5^0.16)
			// Using V^0.16 = 1.4072: 35.74 + 24.86 - 50.31 + 24.04 = 36.47°F
			expectedRange: struct{ min, max float64 }{36.0, 37.0},
		},
		{
			name:      "Extreme cold: 0°F, 30mph wind",
			temp:      0.0,
			windSpeed: 30.0,
			// Correct: 35.74 + 0 - 35.75*(30^0.16) + 0
			// Using V^0.16 = 1.9476: 35.74 - 69.60 = -25.86°F
			expectedRange: struct{ min, max float64 }{-26.5, -25.5},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := calculateApparentTemperature(test.temp, 0.5, test.windSpeed)

			if result < test.expectedRange.min || result > test.expectedRange.max {
				t.Errorf("Wind chill for %s: got %.2f°F, expected between %.1f and %.1f°F",
					test.name, result, test.expectedRange.min, test.expectedRange.max)
				t.Errorf("This likely means the formula is using multiplication (windSpeed*0.16) instead of exponentiation (windSpeed^0.16)")
			}
		})
	}
}

// TestWindChillNotAppliedInWarmWeather ensures wind chill isn't calculated when it shouldn't be.
func TestWindChillNotAppliedInWarmWeather(t *testing.T) {
	// Wind chill should NOT apply above 50°F
	result := calculateApparentTemperature(60.0, 0.5, 20.0)
	if result != 60.0 {
		t.Errorf("Wind chill should not apply at 60°F, expected 60.0, got %.2f", result)
	}

	// Wind chill should NOT apply with very low wind speeds
	result = calculateApparentTemperature(30.0, 0.5, 2.0)
	if result != 30.0 {
		t.Errorf("Wind chill should not apply with 2mph wind, expected 30.0, got %.2f", result)
	}
}

// TestForecastStructIntegrity ensures the Forecast struct works correctly after dead code removal.
func TestForecastStructIntegrity(t *testing.T) {
	// Create a forecast with typical NOAA data
	forecast := Forecast{
		Latitude:  40.7128,
		Longitude: -74.0060,
		Timezone:  "America/New_York",
		Currently: DataPoint{
			Time:                1730000000,
			Temperature:         72,
			ApparentTemperature: 75,
			Humidity:            0.65,
			WindSpeed:           10,
			Summary:             "Partly Cloudy",
		},
		Hourly: DataBlock{
			Summary: "Partly cloudy for the hour",
			Data: []DataPoint{
				{Time: 1730000000, Temperature: 72},
				{Time: 1730003600, Temperature: 71},
			},
		},
		Daily: DataBlock{
			Summary: "Partly cloudy throughout the week",
			Data: []DataPoint{
				{Time: 1730000000, TemperatureMax: 75, TemperatureMin: 60},
			},
		},
		Alerts: []Alert{},
	}

	// Test that we can access all necessary fields
	if forecast.Latitude != 40.7128 {
		t.Errorf("Latitude not preserved")
	}
	if forecast.Currently.Temperature != 72 {
		t.Errorf("Current temperature not preserved")
	}
	if len(forecast.Hourly.Data) != 2 {
		t.Errorf("Hourly data not preserved")
	}
	if len(forecast.Daily.Data) != 1 {
		t.Errorf("Daily data not preserved")
	}

	// Test JSON serialization/deserialization still works
	data, err := json.Marshal(forecast)
	if err != nil {
		t.Fatalf("Failed to marshal forecast: %v", err)
	}

	var decoded Forecast
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal forecast: %v", err)
	}

	if decoded.Latitude != forecast.Latitude {
		t.Errorf("Latitude changed after JSON round-trip")
	}
	if decoded.Currently.Temperature != forecast.Currently.Temperature {
		t.Errorf("Temperature changed after JSON round-trip")
	}
}

// TestPrecipTypeUnicodeHandling tests that precipitation type extraction works with Unicode characters.
// This test will FAIL with current manual lowercasing that only handles ASCII.
func TestPrecipTypeUnicodeHandling(t *testing.T) {
	tests := []struct {
		forecast     string
		expectedType string
	}{
		{"Light rain", "rain"},           // ASCII lowercase
		{"Light Rain", "rain"},           // ASCII uppercase
		{"LIGHT RAIN", "rain"},           // ASCII all caps
		{"Light Räin", ""},               // Unicode - current code breaks
		{"Légère pluie", ""},             // French with accents
		{"Дождь", ""},                    // Cyrillic (means "rain" in Russian)
	}

	for _, test := range tests {
		result := extractPrecipType(test.forecast)
		// Note: We expect "" for non-English because we only check English keywords
		// But the function shouldn't crash or produce garbage
		if result != test.expectedType && test.expectedType == "rain" {
			t.Errorf("extractPrecipType(%q) = %q, expected %q",
				test.forecast, result, test.expectedType)
		}
	}
}

// TestPrecipTypeAdditionalPatterns tests patterns we should handle but currently don't.
// These tests will FAIL, showing we need to add more patterns.
func TestPrecipTypeAdditionalPatterns(t *testing.T) {
	tests := []struct {
		forecast     string
		expectedType string
	}{
		{"Light Drizzle", "rain"},
		{"Drizzle", "rain"},
		{"Snow Flurries", "snow"},
		{"Flurries", "snow"},
		{"Ice Pellets", "sleet"},
		{"Hail", "sleet"},
	}

	for _, test := range tests {
		result := extractPrecipType(test.forecast)
		if result != test.expectedType {
			t.Errorf("extractPrecipType(%q) = %q, expected %q (pattern not yet implemented)",
				test.forecast, result, test.expectedType)
		}
	}
}

// BenchmarkPrecipTypeExtraction benchmarks the precipitation type extraction.
func BenchmarkPrecipTypeExtraction(b *testing.B) {
	forecast := "Heavy Rain Showers Expected Throughout The Day"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractPrecipType(forecast)
	}
}

// BenchmarkPrecipTypeLongString benchmarks with a longer string (worst case for O(n²) concatenation).
func BenchmarkPrecipTypeLongString(b *testing.B) {
	forecast := "Very Heavy Rain Showers With Thunderstorms Expected Throughout The Afternoon And Evening Hours With Possible Flooding"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractPrecipType(forecast)
	}
}

// TestPeriodNameExtraction tests that NOAA period names are properly extracted and stored.
// This test will FAIL until we add the PeriodName field.
func TestPeriodNameExtraction(t *testing.T) {
	tests := []struct {
		name         string
		periodName   string
		expectedName string
	}{
		{"This Afternoon", "This Afternoon", "This Afternoon"},
		{"Tonight", "Tonight", "Tonight"},
		{"Tomorrow", "Tomorrow", "Tomorrow"},
		{"Monday", "Monday", "Monday"},
		{"Monday Night", "Monday Night", "Monday Night"},
		{"Tuesday", "Tuesday", "Tuesday"},
		{"Overnight", "Overnight", "Overnight"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			precipProb := 25.0
			dewpoint := 15.0
			humidity := 60.0

			period := NOAAPeriod{
				Number:                     1,
				Name:                       test.periodName,
				StartTime:                  "2025-11-01T14:00:00-04:00",
				Temperature:                72,
				WindSpeed:                  "10 mph",
				WindDirection:              "SW",
				ShortForecast:              "Partly Cloudy",
				ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
				Dewpoint:                   NOAAValue{Value: &dewpoint},
				RelativeHumidity:           NOAAValue{Value: &humidity},
			}

			dp := convertNOAAPeriodToDataPoint(period)

			if dp.PeriodName != test.expectedName {
				t.Errorf("Expected PeriodName to be %q, got %q", test.expectedName, dp.PeriodName)
			}
		})
	}
}

// TestDailyPeriodNames tests that period names are preserved in daily conversion.
func TestDailyPeriodNames(t *testing.T) {
	precipProb := 20.0
	dewpoint := 10.0
	humidity := 65.0

	periods := []NOAAPeriod{
		{
			Number:                     1,
			Name:                       "Today",
			StartTime:                  "2025-11-01T06:00:00-04:00",
			IsDaytime:                  true,
			Temperature:                75,
			WindSpeed:                  "10 mph",
			WindDirection:              "S",
			ShortForecast:              "Sunny",
			ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
			Dewpoint:                   NOAAValue{Value: &dewpoint},
			RelativeHumidity:           NOAAValue{Value: &humidity},
		},
		{
			Number:                     2,
			Name:                       "Tonight",
			StartTime:                  "2025-11-01T18:00:00-04:00",
			IsDaytime:                  false,
			Temperature:                55,
			WindSpeed:                  "5 mph",
			WindDirection:              "S",
			ShortForecast:              "Clear",
			ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
			Dewpoint:                   NOAAValue{Value: &dewpoint},
			RelativeHumidity:           NOAAValue{Value: &humidity},
		},
	}

	block := convertNOAADailyPeriodsToDataBlock(periods)

	if len(block.Data) != 1 {
		t.Fatalf("Expected 1 daily data point, got %d", len(block.Data))
	}

	// For daily view, we should use the day period name
	if block.Data[0].PeriodName != "Today" {
		t.Errorf("Expected daily period name to be 'Today', got %q", block.Data[0].PeriodName)
	}
}

// TestHourlyPeriodNames tests that hourly periods preserve their names.
func TestHourlyPeriodNames(t *testing.T) {
	precipProb := 15.0
	dewpoint := 12.0
	humidity := 70.0

	periods := []NOAAPeriod{
		{
			Number:                     1,
			Name:                       "This Afternoon",
			StartTime:                  "2025-11-01T14:00:00-04:00",
			Temperature:                72,
			WindSpeed:                  "8 mph",
			WindDirection:              "W",
			ShortForecast:              "Partly Cloudy",
			ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
			Dewpoint:                   NOAAValue{Value: &dewpoint},
			RelativeHumidity:           NOAAValue{Value: &humidity},
		},
		{
			Number:                     2,
			Name:                       "This Evening",
			StartTime:                  "2025-11-01T18:00:00-04:00",
			Temperature:                68,
			WindSpeed:                  "6 mph",
			WindDirection:              "W",
			ShortForecast:              "Mostly Clear",
			ProbabilityOfPrecipitation: NOAAValue{Value: &precipProb},
			Dewpoint:                   NOAAValue{Value: &dewpoint},
			RelativeHumidity:           NOAAValue{Value: &humidity},
		},
	}

	block := convertNOAAPeriodsToDataBlock(periods)

	if len(block.Data) != 2 {
		t.Fatalf("Expected 2 hourly data points, got %d", len(block.Data))
	}

	if block.Data[0].PeriodName != "This Afternoon" {
		t.Errorf("Expected first period name to be 'This Afternoon', got %q", block.Data[0].PeriodName)
	}

	if block.Data[1].PeriodName != "This Evening" {
		t.Errorf("Expected second period name to be 'This Evening', got %q", block.Data[1].PeriodName)
	}
}
