// This package contains the data structures and utilities for retrieving weather forecasts
// from the NOAA National Weather Service API.
package weather

import (
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"log"
	"time"
)

type Flags struct {
	DarkSkyUnavailable string   `json:"darksky-unavailable"`
	DarkSkyStation     string   `json:"datapoint-stations"`
	ISDStations        []string `json:"isds-stations"`
	LAMPStations       []string `json:"lamp-stations"`
	METARStations      []string `json:"metars-stations"`
	METNOLicense       string   `json:"metnol-license"`
	Sources            []string `json:"sources"`
	Units              string   `json:"units"`
}

type DataPoint struct {
	Time                   float64 `json:"time"`
	Summary                string  `json:"summary"`
	Icon                   string  `json:"icon"`
	SunriseTime            float64 `json:"sunriseTime"`
	SunsetTime             float64 `json:"sunsetTime"`
	PrecipIntensity        float64 `json:"precipIntensity"`
	PrecipIntensityMax     float64 `json:"precipIntensityMax"`
	PrecipIntensityMaxTime float64 `json:"precipIntensityMaxTime"`
	PrecipProbability      float64 `json:"precipProbability"`
	PrecipType             string  `json:"precipType"`
	PrecipAccumulation     float64 `json:"precipAccumulation"`
	Temperature            float64 `json:"temperature"`
	TemperatureMin         float64 `json:"temperatureMin"`
	TemperatureMinTime     float64 `json:"temperatureMinTime"`
	TemperatureMax         float64 `json:"temperatureMax"`
	TemperatureMaxTime     float64 `json:"temperatureMaxTime"`
	ApparentTemperature    float64 `json:"apparentTemperature"`
	DewPoint               float64 `json:"dewPoint"`
	WindSpeed              float64 `json:"windSpeed"`
	WindBearing            float64 `json:"windBearing"`
	CloudCover             float64 `json:"cloudCover"`
	Humidity               float64 `json:"humidity"`
	Pressure               float64 `json:"pressure"`
	Visibility             float64 `json:"visibility"`
	Ozone                  float64 `json:"ozone"`
	MoonPhase              float64 `json:"moonPhase"`
	UVIndex                float64 `json:"uvIndex"`
	UVIndexTime            float64 `json:"uvIndexTime"`
}

type DataBlock struct {
	Summary string      `json:"summary"`
	Icon    string      `json:"icon"`
	Data    []DataPoint `json:"data"`
}

type Alert struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Time        float64 `json:"time"`
	Expires     float64 `json:"expires"`
	URI         string  `json:"uri"`
}

type Forecast struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timezone  string    `json:"timezone"`
	Offset    float64   `json:"offset"`
	Currently DataPoint `json:"currently"`
	Minutely  DataBlock `json:"minutely"`
	Hourly    DataBlock `json:"hourly"`
	Daily     DataBlock `json:"daily"`
	Alerts    []Alert   `json:"alerts"`
	Flags     Flags     `json:"flags"`
	APICalls  int       `json:"apicalls"`
	Code      int       `json:"code"`
}

type Units string

const (
	CA   Units = "ca"
	SI   Units = "si"
	US   Units = "us"
	UK   Units = "uk"
	AUTO Units = "auto"
)

const DarkSkyAddress = "https://api.darksky.net/forecast/"
const DarkSkyUnits = "auto"

// NOAA API constants
const NOAABaseURL = "https://api.weather.gov"
const NOAAUserAgent = "vaporwair/2.0 (https://github.com/jeff-bruemmer/vaporwair)"

// NOAA API response structures
type NOAAPointsResponse struct {
	Properties NOAAPointsProperties `json:"properties"`
}

type NOAAPointsProperties struct {
	GridID          string `json:"gridId"`
	GridX           int    `json:"gridX"`
	GridY           int    `json:"gridY"`
	Forecast        string `json:"forecast"`
	ForecastHourly  string `json:"forecastHourly"`
	RelativeLocation NOAARelativeLocation `json:"relativeLocation"`
}

type NOAARelativeLocation struct {
	Properties NOAARelativeLocationProps `json:"properties"`
}

type NOAARelativeLocationProps struct {
	City  string `json:"city"`
	State string `json:"state"`
}

type NOAAForecastResponse struct {
	Properties NOAAForecastProperties `json:"properties"`
}

type NOAAForecastProperties struct {
	Updated  string        `json:"updated"`
	Periods  []NOAAPeriod  `json:"periods"`
}

type NOAAPeriod struct {
	Number                     int       `json:"number"`
	Name                       string    `json:"name"`
	StartTime                  string    `json:"startTime"`
	EndTime                    string    `json:"endTime"`
	IsDaytime                  bool      `json:"isDaytime"`
	Temperature                int       `json:"temperature"`
	TemperatureUnit            string    `json:"temperatureUnit"`
	TemperatureTrend           *string   `json:"temperatureTrend"` // Nullable
	WindSpeed                  string    `json:"windSpeed"`
	WindGust                   *string   `json:"windGust"` // Nullable
	WindDirection              string    `json:"windDirection"`
	Icon                       string    `json:"icon"`
	ShortForecast              string    `json:"shortForecast"`
	DetailedForecast           string    `json:"detailedForecast"`
	ProbabilityOfPrecipitation NOAAValue `json:"probabilityOfPrecipitation"`
	Dewpoint                   NOAAValue `json:"dewpoint"`
	RelativeHumidity           NOAAValue `json:"relativeHumidity"`
}

type NOAAValue struct {
	UnitCode string   `json:"unitCode"`
	Value    *float64 `json:"value"` // Pointer to handle null values
}

// BuildAirNowURL creates http address for dialer to call Dark Sky API.
func BuildDarkSkyURL(addr string, apikey string, c geolocation.Coordinates, units string) string {
	return addr +
		apikey +
		"/" +
		c.Latitude +
		"," +
		c.Longitude +
		"?units=" +
		units
}

// GetForecast dials the Dark Sky API and returns a Forecast.
func GetForecast(addr string) Forecast {
	var wf Forecast
	// Request coordinates from ip-api and specify timeout in seconds
	// Set gzip bool to false for NOAA API.
	resp, err := dialer.NetReq(addr, 5, false)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Decode JSON response directly into weather forecast.
	err = json.NewDecoder(resp.Body).Decode(&wf)
	if err != nil {
		fmt.Println("Error decoding JSON response from weather API.")
		log.Fatal(err)
	}
	return wf
}

// GetNOAAGridPoint gets the grid coordinates for a given lat/lon from NOAA API.
func GetNOAAGridPoint(c geolocation.Coordinates) (NOAAPointsResponse, error) {
	var points NOAAPointsResponse
	url := fmt.Sprintf("%s/points/%s,%s", NOAABaseURL, c.Latitude, c.Longitude)

	resp, err := dialer.NetReqWithUserAgent(url, 5, false, NOAAUserAgent)
	if err != nil {
		return points, fmt.Errorf("error getting NOAA grid point: %v", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&points)
	if err != nil {
		return points, fmt.Errorf("error decoding NOAA points response: %v", err)
	}

	return points, nil
}

// GetNOAAForecast retrieves forecast from NOAA API using the forecast URL.
func GetNOAAForecast(forecastURL string) (NOAAForecastResponse, error) {
	var forecast NOAAForecastResponse

	resp, err := dialer.NetReqWithUserAgent(forecastURL, 5, false, NOAAUserAgent)
	if err != nil {
		return forecast, fmt.Errorf("error getting NOAA forecast: %v", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&forecast)
	if err != nil {
		return forecast, fmt.Errorf("error decoding NOAA forecast response: %v", err)
	}

	return forecast, nil
}

// ConvertNOAAToForecast converts NOAA API response to our Forecast structure.
func ConvertNOAAToForecast(points NOAAPointsResponse, dailyForecast NOAAForecastResponse, hourlyForecast NOAAForecastResponse, c geolocation.Coordinates) Forecast {
	var forecast Forecast

	forecast.Latitude = parseFloat(c.Latitude)
	forecast.Longitude = parseFloat(c.Longitude)
	forecast.Timezone = "America/New_York" // NOAA doesn't provide timezone, use default

	// Convert current conditions from first hourly period
	if len(hourlyForecast.Properties.Periods) > 0 {
		forecast.Currently = convertNOAAPeriodToDataPoint(hourlyForecast.Properties.Periods[0])
	}

	// Convert hourly forecast
	forecast.Hourly = convertNOAAPeriodsToDataBlock(hourlyForecast.Properties.Periods)

	// Convert daily forecast
	forecast.Daily = convertNOAADailyPeriodsToDataBlock(dailyForecast.Properties.Periods)

	// Minutely forecast is not available from NOAA
	forecast.Minutely = DataBlock{
		Summary: "Minutely forecast not available from NOAA",
		Icon:    "",
		Data:    []DataPoint{},
	}

	return forecast
}

// Helper function to convert NOAA period to DataPoint.
func convertNOAAPeriodToDataPoint(period NOAAPeriod) DataPoint {
	var dp DataPoint

	// Parse time
	t, _ := time.Parse(time.RFC3339, period.StartTime)
	dp.Time = float64(t.Unix())
	dp.Summary = period.ShortForecast
	dp.Icon = mapNOAAIconToIcon(period.ShortForecast)
	dp.Temperature = float64(period.Temperature)

	// Handle precipitation probability (can be null)
	if period.ProbabilityOfPrecipitation.Value != nil {
		dp.PrecipProbability = *period.ProbabilityOfPrecipitation.Value / 100.0
	}

	// Parse wind speed (format: "10 to 15 mph" or "10 mph")
	dp.WindSpeed = parseWindSpeed(period.WindSpeed)
	dp.WindBearing = parseWindDirection(period.WindDirection)

	// Convert dewpoint from Celsius to Fahrenheit if available
	// NOAA API returns dewpoint in Celsius (unitCode: "wmoUnit:degC")
	if period.Dewpoint.Value != nil && *period.Dewpoint.Value != 0 {
		dp.DewPoint = celsiusToFahrenheit(*period.Dewpoint.Value)
	}

	// Humidity (convert from percentage)
	if period.RelativeHumidity.Value != nil {
		dp.Humidity = *period.RelativeHumidity.Value / 100.0
	}

	return dp
}

// Helper function to convert NOAA periods to DataBlock.
func convertNOAAPeriodsToDataBlock(periods []NOAAPeriod) DataBlock {
	var block DataBlock

	if len(periods) > 0 {
		block.Summary = periods[0].DetailedForecast
		block.Icon = mapNOAAIconToIcon(periods[0].ShortForecast)
	}

	block.Data = make([]DataPoint, 0, len(periods))
	for _, period := range periods {
		block.Data = append(block.Data, convertNOAAPeriodToDataPoint(period))
	}

	return block
}

// Helper function to convert daily NOAA periods to DataBlock.
func convertNOAADailyPeriodsToDataBlock(periods []NOAAPeriod) DataBlock {
	var block DataBlock

	if len(periods) > 0 {
		block.Summary = periods[0].DetailedForecast
		block.Icon = mapNOAAIconToIcon(periods[0].ShortForecast)
	}

	// Group periods by day (day and night are separate periods)
	dailyData := make([]DataPoint, 0)
	for i := 0; i < len(periods); i += 2 {
		var dp DataPoint
		dayPeriod := periods[i]

		// Parse time
		t, _ := time.Parse(time.RFC3339, dayPeriod.StartTime)
		dp.Time = float64(t.Unix())
		dp.Summary = dayPeriod.ShortForecast
		dp.Icon = mapNOAAIconToIcon(dayPeriod.ShortForecast)

		// Set temperature max from day period
		dp.TemperatureMax = float64(dayPeriod.Temperature)

		// Set temperature min from night period if available
		if i+1 < len(periods) {
			nightPeriod := periods[i+1]
			dp.TemperatureMin = float64(nightPeriod.Temperature)
		}

		// Handle precipitation probability (can be null)
		if dayPeriod.ProbabilityOfPrecipitation.Value != nil {
			dp.PrecipProbability = *dayPeriod.ProbabilityOfPrecipitation.Value / 100.0
		}

		dp.WindSpeed = parseWindSpeed(dayPeriod.WindSpeed)
		dp.WindBearing = parseWindDirection(dayPeriod.WindDirection)

		// Convert dewpoint from Celsius to Fahrenheit if available
		if dayPeriod.Dewpoint.Value != nil && *dayPeriod.Dewpoint.Value != 0 {
			dp.DewPoint = celsiusToFahrenheit(*dayPeriod.Dewpoint.Value)
		}

		// Humidity (convert from percentage)
		if dayPeriod.RelativeHumidity.Value != nil {
			dp.Humidity = *dayPeriod.RelativeHumidity.Value / 100.0
		}

		dailyData = append(dailyData, dp)
	}

	block.Data = dailyData
	return block
}

// Helper functions for data conversion
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseWindSpeed(s string) float64 {
	// Parse formats like "10 mph" or "10 to 15 mph"
	var speed float64
	fmt.Sscanf(s, "%f", &speed)
	return speed
}

func parseWindDirection(dir string) float64 {
	// Convert direction string to degrees
	directions := map[string]float64{
		"N": 0, "NNE": 22.5, "NE": 45, "ENE": 67.5,
		"E": 90, "ESE": 112.5, "SE": 135, "SSE": 157.5,
		"S": 180, "SSW": 202.5, "SW": 225, "WSW": 247.5,
		"W": 270, "WNW": 292.5, "NW": 315, "NNW": 337.5,
	}

	if bearing, ok := directions[dir]; ok {
		return bearing
	}
	return 0
}

func celsiusToFahrenheit(c float64) float64 {
	return (c * 9.0 / 5.0) + 32.0
}

func mapNOAAIconToIcon(shortForecast string) string {
	// Map NOAA forecast descriptions to icon names
	forecast := shortForecast
	switch {
	case contains(forecast, "Sunny"), contains(forecast, "Clear"):
		return "clear-day"
	case contains(forecast, "Partly Cloudy"), contains(forecast, "Partly Sunny"):
		return "partly-cloudy-day"
	case contains(forecast, "Mostly Cloudy"), contains(forecast, "Cloudy"):
		return "cloudy"
	case contains(forecast, "Rain"), contains(forecast, "Showers"):
		return "rain"
	case contains(forecast, "Snow"):
		return "snow"
	case contains(forecast, "Thunderstorm"):
		return "thunderstorm"
	case contains(forecast, "Fog"):
		return "fog"
	default:
		return "partly-cloudy-day"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetNOAAWeatherForecast is the main function to get weather forecast from NOAA API.
func GetNOAAWeatherForecast(c geolocation.Coordinates) (Forecast, error) {
	// Step 1: Get grid point information
	points, err := GetNOAAGridPoint(c)
	if err != nil {
		return Forecast{}, err
	}

	// Step 2: Get daily forecast
	dailyForecast, err := GetNOAAForecast(points.Properties.Forecast)
	if err != nil {
		return Forecast{}, err
	}

	// Step 3: Get hourly forecast
	hourlyForecast, err := GetNOAAForecast(points.Properties.ForecastHourly)
	if err != nil {
		return Forecast{}, err
	}

	// Step 4: Convert to our Forecast structure
	forecast := ConvertNOAAToForecast(points, dailyForecast, hourlyForecast, c)

	return forecast, nil
}
