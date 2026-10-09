// This package contains the data structures and utilities for retrieving weather forecasts
// from the NOAA National Weather Service API.
package weather

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
)

// DataPoint represents a weather data point for a specific time.
// Fields are populated from NOAA National Weather Service API data.
type DataPoint struct {
	Time                float64 `json:"time"`
	PeriodName          string  `json:"periodName"` // Human-readable period label: "This Afternoon", "Tonight", "Monday", etc.
	Summary             string  `json:"summary"`
	PrecipProbability   float64 `json:"precipProbability"`
	PrecipType          string  `json:"precipType"`
	Temperature         float64 `json:"temperature"`
	TemperatureMin      float64 `json:"temperatureMin"`
	TemperatureMax      float64 `json:"temperatureMax"`
	TemperatureTrend    string  `json:"temperatureTrend"`    // "rising", "falling", "steady"
	ApparentTemperature float64 `json:"apparentTemperature"` // Heat index or wind chill
	DewPoint            float64 `json:"dewPoint"`
	WindSpeed           float64 `json:"windSpeed"`
	WindGust            float64 `json:"windGust"`
	WindBearing         float64 `json:"windBearing"` // Degrees (0-360)
	Humidity            float64 `json:"humidity"`    // 0.0 to 1.0
	Pressure            float64 `json:"pressure"`    // Inches of mercury
	Visibility          float64 `json:"visibility"`  // Miles
	DetailedForecast    string  `json:"detailedForecast"`
	// Night marks an overnight daily period (NOAA's isDaytime is false). It has no daytime high.
	Night bool `json:"night,omitempty"`
}

type DataBlock struct {
	Data []DataPoint `json:"data"`
}

type Alert struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Time        float64 `json:"time"`
	Expires     float64 `json:"expires"`
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

const NOAABaseURL = "https://api.weather.gov"

// noaaTimeout bounds each NOAA request the forecast needs.
const noaaTimeout = 10 * time.Second

// noaaOptionalTimeout bounds the alerts and observation requests, which a forecast can do
// without. The observation takes two in a row, so it still finishes within noaaTimeout.
const noaaOptionalTimeout = 5 * time.Second

// NOAA Alerts API structures
type NOAAAlertResponse struct {
	Features []NOAAAlertFeature `json:"features"`
}

type NOAAAlertFeature struct {
	Properties NOAAAlertProperties `json:"properties"`
}

type NOAAAlertProperties struct {
	Event       string `json:"event"`
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
	Forecast            string `json:"forecast"`
	ForecastHourly      string `json:"forecastHourly"`
	ObservationStations string `json:"observationStations"`
}

type NOAAForecastResponse struct {
	Properties NOAAForecastProperties `json:"properties"`
}

type NOAAForecastProperties struct {
	Periods []NOAAPeriod `json:"periods"`
}

type NOAAPeriod struct {
	Name                       string    `json:"name"`
	StartTime                  string    `json:"startTime"`
	EndTime                    string    `json:"endTime"`
	IsDaytime                  bool      `json:"isDaytime"`
	Temperature                int       `json:"temperature"`
	TemperatureTrend           *string   `json:"temperatureTrend"` // Nullable
	WindSpeed                  string    `json:"windSpeed"`
	WindGust                   *string   `json:"windGust"` // Nullable
	WindDirection              string    `json:"windDirection"`
	ShortForecast              string    `json:"shortForecast"`
	DetailedForecast           string    `json:"detailedForecast"`
	ProbabilityOfPrecipitation NOAAValue `json:"probabilityOfPrecipitation"`
	Dewpoint                   NOAAValue `json:"dewpoint"`
	RelativeHumidity           NOAAValue `json:"relativeHumidity"`
}

type NOAAValue struct {
	Value *float64 `json:"value"` // Pointer to handle null values
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
	if strings.Contains(lowerForecast, "sleet") || strings.Contains(lowerForecast, "ice pellets") || strings.Contains(lowerForecast, "hail") {
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

// GetNOAAGridPoint gets the grid coordinates for a given lat/lon from NOAA API.
func GetNOAAGridPoint(c geolocation.Coordinates) (NOAAPointsResponse, error) {
	var points NOAAPointsResponse
	url := fmt.Sprintf("%s/points/%s,%s", NOAABaseURL, c.Latitude, c.Longitude)

	resp, err := dialer.Get(url, noaaTimeout)
	if err != nil {
		return points, err
	}
	defer resp.Body.Close()

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

	resp, err := dialer.Get(forecastURL, noaaTimeout)
	if err != nil {
		return forecast, err
	}
	defer resp.Body.Close()

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

	if len(forecast.Properties.Periods) == 0 {
		return forecast, fmt.Errorf("NOAA returned empty forecast data")
	}

	return forecast, nil
}

// getTimezoneFromCoordinates estimates timezone based on longitude for US locations.
// This is a simplified approximation - a full implementation would use a timezone database.
func getTimezoneFromCoordinates(lat, lon float64) string {
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
func ConvertNOAAToForecast(dailyForecast NOAAForecastResponse, hourlyForecast NOAAForecastResponse, c geolocation.Coordinates) Forecast {
	var forecast Forecast

	forecast.Latitude = parseFloat(c.Latitude)
	forecast.Longitude = parseFloat(c.Longitude)
	// Estimate timezone from coordinates (NOAA doesn't provide timezone)
	forecast.Timezone = getTimezoneFromCoordinates(forecast.Latitude, forecast.Longitude)

	if len(hourlyForecast.Properties.Periods) > 0 {
		forecast.Currently = convertNOAAPeriodToDataPoint(hourlyForecast.Properties.Periods[0])
	}

	forecast.Hourly = convertNOAAPeriodsToDataBlock(hourlyForecast.Properties.Periods)
	forecast.Daily = convertNOAADailyPeriodsToDataBlock(dailyForecast.Properties.Periods)

	// A "Tonight" entry has no daytime high; use the warmest hour of tonight not yet over.
	if periods := dailyForecast.Properties.Periods; len(periods) > 0 && !periods[0].IsDaytime {
		if end, err := time.Parse(time.RFC3339, periods[0].EndTime); err == nil {
			hourStart := float64(time.Now().Add(-time.Hour).Unix())
			for _, h := range forecast.Hourly.Data {
				if h.Time > hourStart && h.Time < float64(end.Unix()) && h.Temperature > forecast.Daily.Data[0].TemperatureMax {
					forecast.Daily.Data[0].TemperatureMax = h.Temperature
				}
			}
		}
	}

	return forecast
}

// convertNOAAPeriodToDataPoint converts one NOAA period, hourly or daily, to a DataPoint.
func convertNOAAPeriodToDataPoint(period NOAAPeriod) DataPoint {
	var dp DataPoint

	t, _ := time.Parse(time.RFC3339, period.StartTime)
	dp.Time = float64(t.Unix())
	dp.PeriodName = period.Name
	dp.Summary = period.ShortForecast
	dp.DetailedForecast = period.DetailedForecast
	dp.Temperature = float64(period.Temperature)
	dp.TemperatureTrend = getTemperatureTrend(period.TemperatureTrend)

	if period.ProbabilityOfPrecipitation.Value != nil {
		dp.PrecipProbability = *period.ProbabilityOfPrecipitation.Value / 100.0
	}
	dp.PrecipType = extractPrecipType(period.ShortForecast)
	if dp.PrecipType == "" {
		dp.PrecipType = extractPrecipType(period.DetailedForecast)
	}

	dp.WindSpeed = parseFloat(period.WindSpeed) // "10 mph", or "10 to 15 mph" as 10
	dp.WindGust = parseWindGust(period.WindGust)
	dp.WindBearing = parseWindDirection(period.WindDirection)

	// NOAA reports dewpoint in Celsius
	if period.Dewpoint.Value != nil {
		dp.DewPoint = celsiusToFahrenheit(*period.Dewpoint.Value)
	}
	if period.RelativeHumidity.Value != nil {
		dp.Humidity = *period.RelativeHumidity.Value / 100.0
	}

	dp.ApparentTemperature = calculateApparentTemperature(dp.Temperature, dp.Humidity, dp.WindSpeed)
	return dp
}

// parseWindGust extracts wind gust speed from NOAA wind gust string.
// Returns 0 if no gust data is available.
func parseWindGust(gust *string) float64 {
	if gust == nil {
		return 0
	}
	return parseFloat(*gust)
}

// getTemperatureTrend returns temperature trend string if available.
func getTemperatureTrend(trend *string) string {
	if trend == nil {
		return ""
	}
	return *trend
}

func convertNOAAPeriodsToDataBlock(periods []NOAAPeriod) DataBlock {
	data := make([]DataPoint, 0, len(periods))
	for _, period := range periods {
		data = append(data, convertNOAAPeriodToDataPoint(period))
	}
	return DataBlock{Data: data}
}

func convertNOAADailyPeriodsToDataBlock(periods []NOAAPeriod) DataBlock {
	// Group periods by day (day and night are separate periods).
	// In the evening NOAA's first period is "Tonight"; it becomes its own entry so that
	// Data[0] stays today rather than shifting every day forward by one.
	dailyData := make([]DataPoint, 0)
	start := 0
	if len(periods) > 0 && !periods[0].IsDaytime {
		tonight := dailyDataPoint(periods[0])
		tonight.TemperatureMin = float64(periods[0].Temperature)
		dailyData = append(dailyData, tonight)
		start = 1
	}
	// A day whose night is past the end of the forecast has no low, and 0F is a real low,
	// so that day is left out; reports would show it once Tonight is over. A lone day is
	// kept, so Data[0] always exists.
	for i := start; i < len(periods); i += 2 {
		dp := dailyDataPoint(periods[i])
		if i+1 < len(periods) {
			dp.TemperatureMin = float64(periods[i+1].Temperature)
		} else if len(dailyData) > 0 {
			break
		}
		dailyData = append(dailyData, dp)
	}
	return DataBlock{Data: dailyData}
}

// dailyDataPoint builds a daily DataPoint from one NOAA period. TemperatureMax is the
// period's temperature; the caller sets TemperatureMin from the following night.
func dailyDataPoint(p NOAAPeriod) DataPoint {
	dp := convertNOAAPeriodToDataPoint(p)
	dp.TemperatureMax = dp.Temperature
	dp.Night = !p.IsDaytime
	return dp
}

// parseFloat reads the number at the start of s, or returns 0.
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseWindDirection(dir string) float64 {
	directions := map[string]float64{
		"N": 0, "NNE": 22.5, "NE": 45, "ENE": 67.5,
		"E": 90, "ESE": 112.5, "SE": 135, "SSE": 157.5,
		"S": 180, "SSW": 202.5, "SW": 225, "WSW": 247.5,
		"W": 270, "WNW": 292.5, "NW": 315, "NNW": 337.5,
	}
	return directions[dir] // 0 (north) when unknown
}

func celsiusToFahrenheit(c float64) float64 {
	return (c * 9.0 / 5.0) + 32.0
}

// GetNOAAAlerts retrieves active weather alerts for coordinates. Alerts are optional,
// so any failure returns none.
func GetNOAAAlerts(c geolocation.Coordinates) []Alert {
	url := fmt.Sprintf("%s/alerts/active?point=%s,%s", NOAABaseURL, c.Latitude, c.Longitude)

	resp, err := dialer.Get(url, noaaOptionalTimeout)
	if err != nil {
		return []Alert{}
	}
	defer resp.Body.Close()

	var alertResp NOAAAlertResponse
	if resp.StatusCode != 200 || json.NewDecoder(resp.Body).Decode(&alertResp) != nil {
		return []Alert{}
	}

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

		alerts = append(alerts, Alert{
			Title:       props.Event,
			Description: props.Description,
			Time:        onset,
			Expires:     expires,
		})
	}
	return alerts
}

// GetNOAAObservation retrieves current observation data for pressure and visibility.
func GetNOAAObservation(stationsURL string) (NOAAObservationProperties, error) {
	var obsProps NOAAObservationProperties

	resp, err := dialer.Get(stationsURL, noaaOptionalTimeout)
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

	resp, err = dialer.Get(obsURL, noaaOptionalTimeout)
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
	points, err := GetNOAAGridPoint(c)
	if err != nil {
		return Forecast{}, err
	}

	// The rest only need the grid point, so they run at once.
	var (
		wg                            sync.WaitGroup
		dailyForecast, hourlyForecast NOAAForecastResponse
		dailyErr, hourlyErr           error
		alerts                        []Alert
		observation                   NOAAObservationProperties
	)
	wg.Add(4)
	go func() {
		defer wg.Done()
		dailyForecast, dailyErr = GetNOAAForecast(points.Properties.Forecast)
	}()
	go func() {
		defer wg.Done()
		hourlyForecast, hourlyErr = GetNOAAForecast(points.Properties.ForecastHourly)
	}()
	go func() {
		defer wg.Done()
		alerts = GetNOAAAlerts(c)
	}()
	go func() {
		// Observation data for pressure and visibility is optional
		defer wg.Done()
		if points.Properties.ObservationStations != "" {
			observation, _ = GetNOAAObservation(points.Properties.ObservationStations)
		}
	}()
	wg.Wait()
	if err := cmp.Or(dailyErr, hourlyErr); err != nil {
		return Forecast{}, err
	}

	forecast := ConvertNOAAToForecast(dailyForecast, hourlyForecast, c)
	forecast.Alerts = alerts

	// The observation is live, so only today gets pressure and visibility.
	if len(forecast.Daily.Data) > 0 {
		today := &forecast.Daily.Data[0]
		if v := observation.BarometricPressure.Value; v != nil {
			today.Pressure = *v / 3386.389 // Pa to inHg, the unit NWS reports
		}
		if v := observation.Visibility.Value; v != nil {
			today.Visibility = *v * 0.000621371 // meters to miles
		}
	}

	return forecast, nil
}
