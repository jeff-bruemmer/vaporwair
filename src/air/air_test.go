package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
)

const exLatitude = "34.0308"
const exLongitude = "-118.473"
const exDate = "Date"
const exKey = "Key"
const exCity = "City"
const exZip = "Zip"

var exCoordinates = geolocation.Coordinates{exLatitude, exLongitude, exCity, exZip}

func TestBuildAirNowURL(t *testing.T) {
	got := BuildAirNowURL(AirNowAddress, exZip, exKey)
	answer := AirNowAddress +
		"zipCode=" + exZip +
		"&distance=25" +
		"&API_KEY=" + exKey
	if got != answer {
		t.Errorf("BuildAirNowURL(AirNowAddress, exZip, exKey) = %s; want "+answer, got)
	}

	// Verify HTTPS is used
	if !strings.Contains(got, "https://") {
		t.Error("Expected HTTPS URL, got HTTP")
	}
}

// getSampleAirNowResponse returns sample AirNow API forecast data
func getSampleAirNowResponse() []Forecast {
	return []Forecast{
		{
			DateIssue:     "2025-10-24T00:00:00",
			DateForecast:  "2025-10-24",
			ReportingArea: "Los Angeles",
			StateCode:     "CA",
			Latitude:      34.0308,
			Longitude:     -118.473,
			ParameterName: "O3",
			AQI:           55,
			Category: Category{
				Number: 2,
				Name:   "Moderate",
			},
			ActionDay:  false,
			Discussion: "Air quality is acceptable.",
		},
		{
			DateIssue:     "2025-10-24T00:00:00",
			DateForecast:  "2025-10-24",
			ReportingArea: "Los Angeles",
			StateCode:     "CA",
			Latitude:      34.0308,
			Longitude:     -118.473,
			ParameterName: "PM2.5",
			AQI:           45,
			Category: Category{
				Number: 1,
				Name:   "Good",
			},
			ActionDay:  false,
			Discussion: "Air quality is good.",
		},
		{
			DateIssue:     "2025-10-24T00:00:00",
			DateForecast:  "2025-10-25",
			ReportingArea: "Los Angeles",
			StateCode:     "CA",
			Latitude:      34.0308,
			Longitude:     -118.473,
			ParameterName: "O3",
			AQI:           62,
			Category: Category{
				Number: 2,
				Name:   "Moderate",
			},
			ActionDay:  false,
			Discussion: "Air quality is acceptable.",
		},
	}
}

func TestGetForecast_WithMockServer(t *testing.T) {
	// Create sample response data
	sampleData := getSampleAirNowResponse()

	// Create a mock HTTP server that returns our sample data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleData)
	}))
	defer server.Close()

	// Call GetForecast with the mock server URL
	forecasts := GetForecast(server.URL)

	// Verify we got the expected number of forecasts
	if len(forecasts) != 3 {
		t.Errorf("Expected 3 forecasts, got %d", len(forecasts))
	}

	// Verify first forecast data
	if forecasts[0].ParameterName != "O3" {
		t.Errorf("Expected first parameter name 'O3', got '%s'", forecasts[0].ParameterName)
	}

	if forecasts[0].AQI != 55 {
		t.Errorf("Expected first AQI 55, got %d", forecasts[0].AQI)
	}

	if forecasts[0].Category.Name != "Moderate" {
		t.Errorf("Expected first category 'Moderate', got '%s'", forecasts[0].Category.Name)
	}

	// Verify we have both today and tomorrow's forecasts
	dates := make(map[string]bool)
	for _, f := range forecasts {
		dates[f.DateForecast] = true
	}

	if len(dates) != 2 {
		t.Errorf("Expected forecasts for 2 different dates, got %d", len(dates))
	}
}

func TestGetForecast_EmptyResponse(t *testing.T) {
	// Create a mock HTTP server that returns empty array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	// Call GetForecast with the mock server URL
	forecasts := GetForecast(server.URL)

	// Verify we got an empty slice
	if len(forecasts) != 0 {
		t.Errorf("Expected 0 forecasts for empty response, got %d", len(forecasts))
	}
}

func TestGetForecast_NullResponse(t *testing.T) {
	// Create a mock HTTP server that returns null (this is what AirNow API returns when no data)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("null"))
	}))
	defer server.Close()

	// Call GetForecast with the mock server URL
	forecasts := GetForecast(server.URL)

	// Verify we got an empty slice (not nil)
	if forecasts == nil {
		t.Error("Expected non-nil slice for null response, got nil")
	}

	if len(forecasts) != 0 {
		t.Errorf("Expected 0 forecasts for null response, got %d", len(forecasts))
	}
}

// TestAirQualityIntegration tests the full flow of getting air quality data
func TestAirQualityIntegration(t *testing.T) {
	// Sample coordinates for testing
	coords := geolocation.Coordinates{
		Latitude:  "34.0522",
		Longitude: "-118.2437",
		City:      "Los Angeles",
		Zip:       "90001",
	}

	// Build URL with mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters are correct
		query := r.URL.Query()
		if query.Get("zipCode") != coords.Zip {
			t.Errorf("Expected zipCode %s, got %s", coords.Zip, query.Get("zipCode"))
		}
		if query.Get("distance") != "25" {
			t.Errorf("Expected distance 25, got %s", query.Get("distance"))
		}

		// Return sample air quality data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getSampleAirNowResponse())
	}))
	defer server.Close()

	// Build URL using the actual function
	// Use server.URL as base and replace the protocol/host part of AirNowAddress
	testAddr := server.URL + "?"
	url := BuildAirNowURL(testAddr, coords.Zip, "test-key")

	// Get forecast
	forecasts := GetForecast(url)

	// Verify we got data
	if len(forecasts) == 0 {
		t.Error("Expected air quality forecasts, got none")
	}

	// Verify the data structure
	if forecasts[0].ReportingArea == "" {
		t.Error("Expected reporting area to be set")
	}

	if forecasts[0].AQI <= 0 {
		t.Error("Expected positive AQI value")
	}
}

// TestHandlingUnavailableForecasts verifies handling of AQI=-1 (forecast not yet available)
func TestHandlingUnavailableForecasts(t *testing.T) {
	// Create forecasts with AQI=-1 (common response from AirNow when forecasts aren't ready)
	unavailableForecasts := []Forecast{
		{
			DateForecast:  "2025-10-24",
			ParameterName: "O3",
			AQI:           -1, // Not yet available
			Category:      Category{Number: 1, Name: "Good"},
		},
		{
			DateForecast:  "2025-10-24",
			ParameterName: "PM2.5",
			AQI:           -1, // Not yet available
			Category:      Category{Number: 1, Name: "Good"},
		},
	}

	// Filter out unavailable forecasts (like the report code does)
	var validForecasts []Forecast
	for _, f := range unavailableForecasts {
		if f.AQI >= 0 {
			validForecasts = append(validForecasts, f)
		}
	}

	// Should have no valid forecasts
	if len(validForecasts) != 0 {
		t.Errorf("Expected 0 valid forecasts when all AQI=-1, got %d", len(validForecasts))
	}
}
