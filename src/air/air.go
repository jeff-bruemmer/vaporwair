// This package contains the data structures and facilities for retrieving forecasts
// from the AirNow API
package air

import (
	"encoding/json"
	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
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
		log.Fatal(err)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&af)

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

const AirNowAddress = "https://www.airnowapi.org/aq/forecast/latLong/?format=application/json&"

// BuildAirNowURL creates http address for dialer to call Air Now API.
func BuildAirNowURL(addr string, c geolocation.Coordinates, date string, apiKey string) string {
	return addr +
		"latitude=" + c.Latitude +
		"&longitude=" + c.Longitude +
		"&date=" + date +
		"&distance=25" +
		"&API_KEY=" + apiKey
}

// GetForecast dials AirNow API and returns a slice of Forecasts.
// Returns an empty slice (not nil) if no forecasts are available.
// Uses the default provider for backward compatibility.
func GetForecast(addr string) []Forecast {
	return DefaultProvider.GetForecast(addr)
}
