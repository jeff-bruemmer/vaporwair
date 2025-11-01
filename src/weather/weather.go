// This package contains the data structures and utilities for retrieving weather forecasts
// from the NOAA National Weather Service API.
package weather

import (
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"math"
	"strings"
	"time"
)

// DataPoint represents a weather data point for a specific time.
// Fields are populated from NOAA National Weather Service API data.
type DataPoint struct {
	Time                float64 `json:"time"`
	PeriodName          string  `json:"periodName"`          // Human-readable period label: "This Afternoon", "Tonight", "Monday", etc.
	Summary             string  `json:"summary"`
	Icon                string  `json:"icon"`
	PrecipProbability   float64 `json:"precipProbability"`
	PrecipType          string  `json:"precipType"`
	Temperature         float64 `json:"temperature"`
	TemperatureMin      float64 `json:"temperatureMin"`
	TemperatureMax      float64 `json:"temperatureMax"`
	TemperatureTrend    string  `json:"temperatureTrend"` // "rising", "falling", "steady"
	ApparentTemperature float64 `json:"apparentTemperature"` // Heat index or wind chill
	DewPoint            float64 `json:"dewPoint"`
	WindSpeed           float64 `json:"windSpeed"`
	WindGust            float64 `json:"windGust"`
	WindBearing         float64 `json:"windBearing"` // Degrees (0-360)
	CloudCover          float64 `json:"cloudCover"`  // 0.0 to 1.0
	Humidity            float64 `json:"humidity"`    // 0.0 to 1.0
	Pressure            float64 `json:"pressure"`    // Atmospheres
	Visibility          float64 `json:"visibility"`  // Miles
	UVIndex             float64 `json:"uvIndex"`     // NOTE: Not provided by NOAA, always 0. Would require separate EPA UV Index API.
	DetailedForecast    string  `json:"detailedForecast"`
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

// Forecast represents a complete weather forecast with current conditions,
// hourly and daily data, and any active weather alerts.
type Forecast struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timezone  string    `json:"timezone"`
	Currently DataPoint `json:"currently"`
	Hourly    DataBlock `json:"hourly"`
	Daily     DataBlock `json:"daily"`
	Alerts    []Alert   `json:"alerts"`
}

// NOAA API constants
const NOAABaseURL = "https://api.weather.gov"
const NOAAUserAgent = "vaporwair/2.0 (https://github.com/jeff-bruemmer/vaporwair)"

// NOAA Alerts API structures
type NOAAAlertResponse struct {
	Features []NOAAAlertFeature `json:"features"`
}

type NOAAAlertFeature struct {
	Properties NOAAAlertProperties `json:"properties"`
}

type NOAAAlertProperties struct {
	Event       string `json:"event"`
	Headline    string `json:"headline"`
	Description string `json:"description"`
	Onset       string `json:"onset"`
	Expires     string `json:"expires"`
}

// NOAA Observation Station structures
type NOAAStationsResponse struct {
	Features []NOAAStationFeature `json:"features"`
}

type NOAAStationFeature struct {
	Properties NOAAStationProperties `json:"properties"`
}

type NOAAStationProperties struct {
	StationIdentifier string `json:"stationIdentifier"`
}

type NOAAObservationResponse struct {
	Properties NOAAObservationProperties `json:"properties"`
}

type NOAAObservationProperties struct {
	BarometricPressure NOAAValue `json:"barometricPressure"`
	Visibility         NOAAValue `json:"visibility"`
}

// NOAA API response structures
type NOAAPointsResponse struct {
	Properties NOAAPointsProperties `json:"properties"`
}

type NOAAPointsProperties struct {
	GridID               string               `json:"gridId"`
	GridX                int                  `json:"gridX"`
	GridY                int                  `json:"gridY"`
	Forecast             string               `json:"forecast"`
	ForecastHourly       string               `json:"forecastHourly"`
	ObservationStations  string               `json:"observationStations"`
	RelativeLocation     NOAARelativeLocation `json:"relativeLocation"`
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

// calculateApparentTemperature calculates "feels like" temperature based on conditions.
// Uses heat index for hot conditions and wind chill for cold conditions.
func calculateApparentTemperature(temp, humidity, windSpeed float64) float64 {
	// For temperatures above 80°F with humidity, use heat index
	if temp >= 80 {
		// Simplified heat index formula
		hi := -42.379 +
			2.04901523*temp +
			10.14333127*humidity*100 -
			0.22475541*temp*humidity*100 -
			0.00683783*temp*temp -
			0.05481717*(humidity*100)*(humidity*100) +
			0.00122874*temp*temp*(humidity*100) +
			0.00085282*temp*(humidity*100)*(humidity*100) -
			0.00000199*temp*temp*(humidity*100)*(humidity*100)

		// If conditions don't warrant heat index, just return temp
		if humidity < 0.40 {
			return temp
		}
		return hi
	}

	// For temperatures below 50°F with wind, use wind chill
	if temp <= 50 && windSpeed > 3 {
		// Wind chill formula (valid for temps ≤ 50°F and wind speeds > 3 mph)
		// Uses NWS formula: WC = 35.74 + 0.6215*T - 35.75*(V^0.16) + 0.4275*T*(V^0.16)
		// where T is temperature in °F and V is wind speed in mph
		windPower := math.Pow(windSpeed, 0.16)
		windChill := 35.74 +
			0.6215*temp -
			35.75*windPower +
			0.4275*temp*windPower
		return windChill
	}

	// For moderate conditions, return actual temperature
	return temp
}

// extractPrecipType determines precipitation type from forecast description.
func extractPrecipType(forecast string) string {
	lowerForecast := strings.ToLower(forecast)

	// Check for mixed precipitation first
	if strings.Contains(lowerForecast, "rain") && strings.Contains(lowerForecast, "snow") {
		return "rain/snow"
	}
	if strings.Contains(lowerForecast, "wintry mix") {
		return "mix"
	}

	// Check specific types (most specific first)
	if strings.Contains(lowerForecast, "freezing rain") {
		return "freezing rain"
	}
	if strings.Contains(lowerForecast, "sleet") || strings.Contains(lowerForecast, "ice pellets") {
		return "sleet"
	}
	if strings.Contains(lowerForecast, "hail") {
		return "sleet"
	}
	if strings.Contains(lowerForecast, "snow") || strings.Contains(lowerForecast, "flurries") {
		return "snow"
	}
	if strings.Contains(lowerForecast, "rain") || strings.Contains(lowerForecast, "showers") || strings.Contains(lowerForecast, "drizzle") {
		return "rain"
	}

	return ""
}

// isIconDaytime determines if NOAA icon URL represents daytime or nighttime.
func isIconDaytime(iconURL string) bool {
	return strings.Contains(iconURL, "/day/")
}

// estimateCloudCover estimates cloud cover percentage from forecast description.
func estimateCloudCover(forecast string) float64 {
	// Check more specific patterns first to avoid false matches
	switch {
	case strings.Contains(forecast, "Cloudy"), strings.Contains(forecast, "Overcast"):
		// Check for "Mostly Cloudy" first
		if strings.Contains(forecast, "Mostly Cloudy") {
			return 0.75
		}
		// Check for "Partly Cloudy" next
		if strings.Contains(forecast, "Partly Cloudy") {
			return 0.50
		}
		// Just "Cloudy" or "Overcast"
		return 1.0
	case strings.Contains(forecast, "Sunny"):
		// Check for "Mostly Sunny" first
		if strings.Contains(forecast, "Mostly Sunny") {
			return 0.25
		}
		// Check for "Partly Sunny" next
		if strings.Contains(forecast, "Partly Sunny") {
			return 0.50
		}
		// Just "Sunny"
		return 0.0
	case strings.Contains(forecast, "Clear"):
		return 0.0
	default:
		return 0.50 // Default to partly cloudy
	}
}

// GetNOAAGridPoint gets the grid coordinates for a given lat/lon from NOAA API.
func GetNOAAGridPoint(c geolocation.Coordinates) (NOAAPointsResponse, error) {
	var points NOAAPointsResponse
	url := fmt.Sprintf("%s/points/%s,%s", NOAABaseURL, c.Latitude, c.Longitude)

	resp, err := dialer.NetReqWithUserAgent(url, 10, false, NOAAUserAgent)
	if err != nil {
		return points, fmt.Errorf("failed to connect to NOAA weather service at %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode == 404 {
		return points, fmt.Errorf("NOAA does not have weather data for coordinates %s, %s (location may be outside US coverage)", c.Latitude, c.Longitude)
	}
	if resp.StatusCode >= 500 {
		return points, fmt.Errorf("NOAA weather service is experiencing issues (status %d) - please try again later", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return points, fmt.Errorf("NOAA weather service returned error status %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&points)
	if err != nil {
		return points, fmt.Errorf("failed to parse NOAA grid point response: %w", err)
	}

	return points, nil
}

// GetNOAAForecast retrieves forecast from NOAA API using the forecast URL.
func GetNOAAForecast(forecastURL string) (NOAAForecastResponse, error) {
	var forecast NOAAForecastResponse

	resp, err := dialer.NetReqWithUserAgent(forecastURL, 10, false, NOAAUserAgent)
	if err != nil {
		return forecast, fmt.Errorf("failed to retrieve forecast from NOAA: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode == 404 {
		return forecast, fmt.Errorf("NOAA forecast endpoint not found (this may indicate a service change)")
	}
	if resp.StatusCode >= 500 {
		return forecast, fmt.Errorf("NOAA weather service is experiencing issues (status %d) - please try again later", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return forecast, fmt.Errorf("NOAA weather service returned error status %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&forecast)
	if err != nil {
		return forecast, fmt.Errorf("failed to parse NOAA forecast response: %w", err)
	}

	// Validate we got forecast data
	if len(forecast.Properties.Periods) == 0 {
		return forecast, fmt.Errorf("NOAA returned empty forecast data")
	}

	return forecast, nil
}

// getTimezoneFromCoordinates estimates timezone based on longitude for US locations.
// This is a simplified approximation - a full implementation would use a timezone database.
func getTimezoneFromCoordinates(lat, lon float64) string {
	// Simple timezone estimation for continental US based on longitude
	// Eastern: > -87.5°
	// Central: -87.5° to -101.5°
	// Mountain: -101.5° to -115°
	// Pacific: < -115°

	switch {
	case lon > -87.5:
		return "America/New_York" // Eastern
	case lon > -101.5:
		return "America/Chicago" // Central
	case lon > -115:
		return "America/Denver" // Mountain
	case lon > -125:
		return "America/Los_Angeles" // Pacific
	default:
		// Alaska, Hawaii, or outside continental US
		if lat > 50 {
			return "America/Anchorage" // Alaska
		} else if lat < 25 {
			return "Pacific/Honolulu" // Hawaii
		}
		return "America/New_York" // Default fallback
	}
}

// ConvertNOAAToForecast converts NOAA API response to our Forecast structure.
func ConvertNOAAToForecast(points NOAAPointsResponse, dailyForecast NOAAForecastResponse, hourlyForecast NOAAForecastResponse, c geolocation.Coordinates) Forecast {
	var forecast Forecast

	forecast.Latitude = parseFloat(c.Latitude)
	forecast.Longitude = parseFloat(c.Longitude)
	// Estimate timezone from coordinates (NOAA doesn't provide timezone)
	forecast.Timezone = getTimezoneFromCoordinates(forecast.Latitude, forecast.Longitude)

	// Convert current conditions from first hourly period
	if len(hourlyForecast.Properties.Periods) > 0 {
		forecast.Currently = convertNOAAPeriodToDataPoint(hourlyForecast.Properties.Periods[0])
	}

	// Convert hourly forecast
	forecast.Hourly = convertNOAAPeriodsToDataBlock(hourlyForecast.Properties.Periods)

	// Convert daily forecast
	forecast.Daily = convertNOAADailyPeriodsToDataBlock(dailyForecast.Properties.Periods)

	// Note: NOAA does not provide minutely forecasts (only hourly and daily)

	return forecast
}

// Helper function to convert NOAA period to DataPoint.
func convertNOAAPeriodToDataPoint(period NOAAPeriod) DataPoint {
	var dp DataPoint

	// Parse time
	t, _ := time.Parse(time.RFC3339, period.StartTime)
	dp.Time = float64(t.Unix())
	dp.PeriodName = period.Name
	dp.Summary = period.ShortForecast
	dp.DetailedForecast = period.DetailedForecast
	dp.Icon = mapNOAAIconToIcon(period.ShortForecast)
	dp.Temperature = float64(period.Temperature)
	dp.TemperatureTrend = getTemperatureTrend(period.TemperatureTrend)

	// Handle precipitation probability (can be null)
	if period.ProbabilityOfPrecipitation.Value != nil {
		dp.PrecipProbability = *period.ProbabilityOfPrecipitation.Value / 100.0
	}

	// Extract precipitation type from forecast description
	dp.PrecipType = extractPrecipType(period.ShortForecast)
	if dp.PrecipType == "" && dp.DetailedForecast != "" {
		// Try detailed forecast if short forecast didn't have precip type
		dp.PrecipType = extractPrecipType(period.DetailedForecast)
	}

	// Parse wind speed and gust
	dp.WindSpeed = parseWindSpeed(period.WindSpeed)
	dp.WindGust = parseWindGust(period.WindGust)
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

	// Calculate apparent temperature (feels like)
	dp.ApparentTemperature = calculateApparentTemperature(
		dp.Temperature,
		dp.Humidity,
		dp.WindSpeed,
	)

	// Estimate cloud cover from forecast description
	dp.CloudCover = estimateCloudCover(period.ShortForecast)

	return dp
}

// parseWindGust extracts wind gust speed from NOAA wind gust string.
// Returns 0 if no gust data is available.
func parseWindGust(gust *string) float64 {
	if gust == nil || *gust == "" {
		return 0
	}
	var speed float64
	fmt.Sscanf(*gust, "%f", &speed)
	return speed
}

// getTemperatureTrend returns temperature trend string if available.
func getTemperatureTrend(trend *string) string {
	if trend == nil {
		return ""
	}
	return *trend
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
		var nightPeriod NOAAPeriod
		hasNightPeriod := i+1 < len(periods)
		if hasNightPeriod {
			nightPeriod = periods[i+1]
		}

		// Determine which period is day and which is night based on IsDaytime flag
		if !dayPeriod.IsDaytime {
			// First period is actually night, swap them
			dayPeriod, nightPeriod = nightPeriod, dayPeriod
		}

		// Parse time
		t, _ := time.Parse(time.RFC3339, dayPeriod.StartTime)
		dp.Time = float64(t.Unix())
		dp.PeriodName = dayPeriod.Name
		dp.Summary = dayPeriod.ShortForecast
		dp.DetailedForecast = dayPeriod.DetailedForecast
		dp.Icon = mapNOAAIconToIcon(dayPeriod.ShortForecast)

		// Set temperature max from day period, min from night period
		dp.TemperatureMax = float64(dayPeriod.Temperature)
		dp.TemperatureTrend = getTemperatureTrend(dayPeriod.TemperatureTrend)
		if hasNightPeriod {
			dp.TemperatureMin = float64(nightPeriod.Temperature)
		}

		// Handle precipitation probability (can be null)
		if dayPeriod.ProbabilityOfPrecipitation.Value != nil {
			dp.PrecipProbability = *dayPeriod.ProbabilityOfPrecipitation.Value / 100.0
		}

		dp.WindSpeed = parseWindSpeed(dayPeriod.WindSpeed)
		dp.WindGust = parseWindGust(dayPeriod.WindGust)
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
	// Note: Order matters - check more specific patterns first
	forecast := shortForecast
	switch {
	case strings.Contains(forecast, "Partly Cloudy"), strings.Contains(forecast, "Partly Sunny"):
		return "partly-cloudy-day"
	case strings.Contains(forecast, "Mostly Cloudy"):
		return "cloudy"
	case strings.Contains(forecast, "Cloudy"):
		return "cloudy"
	case strings.Contains(forecast, "Sunny"), strings.Contains(forecast, "Clear"):
		return "clear-day"
	case strings.Contains(forecast, "Rain"), strings.Contains(forecast, "Showers"):
		return "rain"
	case strings.Contains(forecast, "Snow"):
		return "snow"
	case strings.Contains(forecast, "Thunderstorm"):
		return "thunderstorm"
	case strings.Contains(forecast, "Fog"):
		return "fog"
	default:
		return "partly-cloudy-day"
	}
}

// GetNOAAAlerts retrieves active weather alerts for coordinates.
func GetNOAAAlerts(c geolocation.Coordinates) ([]Alert, error) {
	url := fmt.Sprintf("%s/alerts/active?point=%s,%s", NOAABaseURL, c.Latitude, c.Longitude)

	resp, err := dialer.NetReqWithUserAgent(url, 10, false, NOAAUserAgent)
	if err != nil {
		// Alerts are optional - don't fail if unavailable
		return []Alert{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Alerts endpoint error - return empty slice
		return []Alert{}, nil
	}

	var alertResp NOAAAlertResponse
	err = json.NewDecoder(resp.Body).Decode(&alertResp)
	if err != nil {
		return []Alert{}, nil
	}

	// Convert NOAA alerts to our Alert structure
	alerts := make([]Alert, 0)
	for _, feature := range alertResp.Features {
		props := feature.Properties

		var onset, expires float64
		if t, err := time.Parse(time.RFC3339, props.Onset); err == nil {
			onset = float64(t.Unix())
		}
		if t, err := time.Parse(time.RFC3339, props.Expires); err == nil {
			expires = float64(t.Unix())
		}

		alert := Alert{
			Title:       props.Event,
			Description: props.Description,
			Time:        onset,
			Expires:     expires,
			URI:         "", // NOAA doesn't provide direct URI in alerts
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetNOAAObservation retrieves current observation data for pressure and visibility.
func GetNOAAObservation(stationsURL string) (NOAAObservationProperties, error) {
	var obsProps NOAAObservationProperties

	// First, get list of observation stations
	resp, err := dialer.NetReqWithUserAgent(stationsURL, 10, false, NOAAUserAgent)
	if err != nil {
		return obsProps, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return obsProps, fmt.Errorf("observation stations endpoint returned status %d", resp.StatusCode)
	}

	var stations NOAAStationsResponse
	err = json.NewDecoder(resp.Body).Decode(&stations)
	if err != nil {
		return obsProps, err
	}

	if len(stations.Features) == 0 {
		return obsProps, fmt.Errorf("no observation stations found")
	}

	// Get observations from the first (nearest) station
	stationID := stations.Features[0].Properties.StationIdentifier
	obsURL := fmt.Sprintf("%s/stations/%s/observations/latest", NOAABaseURL, stationID)

	resp, err = dialer.NetReqWithUserAgent(obsURL, 10, false, NOAAUserAgent)
	if err != nil {
		return obsProps, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return obsProps, fmt.Errorf("observation endpoint returned status %d", resp.StatusCode)
	}

	var observation NOAAObservationResponse
	err = json.NewDecoder(resp.Body).Decode(&observation)
	if err != nil {
		return obsProps, err
	}

	return observation.Properties, nil
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

	// Step 4: Get weather alerts (optional - don't fail if unavailable)
	alerts, _ := GetNOAAAlerts(c)

	// Step 5: Get observation data for pressure and visibility (optional)
	var observation NOAAObservationProperties
	if points.Properties.ObservationStations != "" {
		observation, _ = GetNOAAObservation(points.Properties.ObservationStations)
	}

	// Step 6: Convert to our Forecast structure
	forecast := ConvertNOAAToForecast(points, dailyForecast, hourlyForecast, c)
	forecast.Alerts = alerts

	// Add observation data to current conditions
	if observation.BarometricPressure.Value != nil {
		// Convert from Pascals to millibars (1 Pa = 0.01 mbar)
		forecast.Currently.Pressure = *observation.BarometricPressure.Value / 100.0
		// Convert to atmospheres (1 atm = 1013.25 mbar)
		forecast.Currently.Pressure = forecast.Currently.Pressure / 1013.25
		// Also add to daily data
		if len(forecast.Daily.Data) > 0 {
			forecast.Daily.Data[0].Pressure = forecast.Currently.Pressure
		}
	}

	if observation.Visibility.Value != nil {
		// Convert from meters to miles (1 meter = 0.000621371 miles)
		forecast.Currently.Visibility = *observation.Visibility.Value * 0.000621371
		// Also add to daily data
		if len(forecast.Daily.Data) > 0 {
			forecast.Daily.Data[0].Visibility = forecast.Currently.Visibility
		}
	}

	return forecast, nil
}
