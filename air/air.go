package air

import (
	"github.com/jeff-bruemmer/vaporwair/dialer"
	"github.com/jeff-bruemmer/vaporwair/geolocation"
	"log"
	"encoding/json"
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

const AirNowAddress = "http://www.airnowapi.org/aq/forecast/latLong/?format=application/json&"

// buildAirNowURL creates http address for dialer to call Air Now API.
func BuildAirNowURL(addr string, c geolocation.Coordinates, date string, apiKey string) string {
	return addr +
		"latitude=" + c.Latitude +
		"&longitude=" + c.Longitude +
		"&date=" + date +
		"&distance=25" +
		"&API_KEY=" + apiKey
}


// GetForecast dials AirNow API and returns a slice of Forecasts.
func GetForecast(addr string) []Forecast {
	var af []Forecast
	resp, err := dialer.NetReq(addr, 10, false)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&af)
	return af
}


