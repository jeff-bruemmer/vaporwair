// This package contains the data structures and facilities for retrieving forecasts
// from the AirNow API
package air

import (
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"io"
	"strings"
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

// apiForecast mirrors the JSON returned by the AirNow current forecast endpoint.
// It is converted to Forecast so report code and cached forecasts keep their shape.
type apiForecast struct {
	DateIssue      string `json:"dateIssue"`
	DateValid      string `json:"dateValid"`
	ReportingArea  string `json:"reportingArea"`
	StateCode      string `json:"stateCode"`
	ParameterName  string `json:"parameterName"`
	AQI            int    `json:"aqi"`
	CategoryNumber int    `json:"categoryNumber"`
	CategoryName   string `json:"categoryName"`
	ActionDay      bool   `json:"actionDay"`
	Discussion     string `json:"discussion"`
}

func (a apiForecast) toForecast() Forecast {
	return Forecast{
		DateIssue:     a.DateIssue,
		DateForecast:  a.DateValid,
		ReportingArea: a.ReportingArea,
		StateCode:     a.StateCode,
		ParameterName: a.ParameterName,
		AQI:           a.AQI,
		Category:      Category{Number: a.CategoryNumber, Name: a.CategoryName},
		ActionDay:     a.ActionDay,
		Discussion:    a.Discussion,
	}
}

const AirNowAddress = "https://www.airnowapi.org/aq/forecast/current/?format=application/json&"

// BuildAirNowURL creates http address for dialer to call Air Now API.
func BuildAirNowURL(addr string, zipCode string, apiKey string) string {
	return addr +
		"zipCode=" + zipCode +
		"&distance=25" +
		"&API_KEY=" + apiKey
}

// GetForecast dials the AirNow current forecast endpoint and returns a slice of Forecasts.
// Errors are returned rather than logged so the caller can report them after the spinner stops.
func GetForecast(addr string) ([]Forecast, error) {
	var af []apiForecast

	resp, err := dialer.NetReq(addr, 10, false)
	if err != nil {
		return []Forecast{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		if msg := strings.TrimSpace(string(body)); msg != "" {
			return []Forecast{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
		}
		return []Forecast{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&af)
	if err != nil {
		return []Forecast{}, fmt.Errorf("could not decode response: %w", err)
	}

	forecasts := make([]Forecast, 0, len(af))
	for _, f := range af {
		forecasts = append(forecasts, f.toForecast())
	}
	return forecasts, nil
}
