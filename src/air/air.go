// This package contains the data structures and facilities for retrieving forecasts
// from the AirNow API
package air

import (
	"encoding/json"
	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"log"
)

// Provider defines an interface for air quality data retrieval.
// This allows for easy mocking in tests.
type Provider interface {
	GetForecast(addr string) []Forecast
}

// AirNowProvider implements the Provider interface for AirNow data.
type AirNowProvider struct{}

// GetForecast retrieves air quality forecast from AirNow API.
func (p *AirNowProvider) GetForecast(addr string) []Forecast {
	var af []Forecast

	resp, err := dialer.NetReq(addr, 10, false)
	if err != nil {
		log.Printf("Warning: AirNow API request failed: %v\n", err)
		return []Forecast{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Warning: AirNow API returned status %d - air quality data unavailable\n", resp.StatusCode)
		return []Forecast{}
	}

	err = json.NewDecoder(resp.Body).Decode(&af)
	if err != nil {
		log.Printf("Warning: Failed to decode AirNow response: %v\n", err)
		return []Forecast{}
	}

	// If API returns null or decode fails, return empty slice instead of nil
	if af == nil {
		return []Forecast{}
	}
	return af
}

// DefaultProvider is the default air quality provider used by the application.
var DefaultProvider Provider = &AirNowProvider{}

type Category struct {
	Number int    `json:"Number"`
	Name   string `json:"Name"`
}

type Forecast struct {
	DateIssue     string   `json:"DateIssue"`
	DateForecast  string   `json:"DateForecast"`
	ReportingArea string   `json:"ReportingArea"`
	StateCode     string   `json:"StateCode"`
	Latitude      float64  `json:"Latitude"`
	Longitude     float64  `json:"Longitude"`
	ParameterName string   `json:"ParameterName"`
	AQI           int      `json:"AQI"`
	Category      Category `json:"Category"`
	ActionDay     bool     `json:"ActionDay"`
	Discussion    string   `json:"Discussion"`
}

const AirNowAddress = "https://www.airnowapi.org/aq/observation/zipCode/current/?format=application/json&"

// BuildAirNowURL creates http address for dialer to call Air Now API.
func BuildAirNowURL(addr string, zipCode string, apiKey string) string {
	return addr +
		"zipCode=" + zipCode +
		"&distance=25" +
		"&API_KEY=" + apiKey
}

// GetForecast dials AirNow API and returns a slice of Forecasts.
// Returns an empty slice (not nil) if no forecasts are available.
// Uses the default provider for backward compatibility.
func GetForecast(addr string) []Forecast {
	return DefaultProvider.GetForecast(addr)
}
