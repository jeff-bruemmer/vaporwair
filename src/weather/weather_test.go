package weather

import (
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
