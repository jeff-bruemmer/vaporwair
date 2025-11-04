// This package contains the data structures and facilities for retrieving forecasts
// from the AirNow API
package air

import (
	"encoding/json"
	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"log"
)

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
// Returns an empty slice if no forecasts are available.
func GetForecast(addr string) []Forecast {
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

	if af == nil {
		return []Forecast{}
	}
	return af
}
