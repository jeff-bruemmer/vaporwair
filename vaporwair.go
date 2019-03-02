package main

import (
	"flag"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/air"
	"github.com/jeff-bruemmer/vaporwair/geolocation"
	"github.com/jeff-bruemmer/vaporwair/report"
	"github.com/jeff-bruemmer/vaporwair/storage"
	"github.com/jeff-bruemmer/vaporwair/weather"
	"time"
)

var weatherHourly bool
var weatherWeek bool
var airQuality bool
var weatherForecast weather.Forecast
var airForecast []air.Forecast
var config storage.Config

// Is the forecast still valid? Specify a timeout duration (in minutes) that determines whether or not
// the forecast is still valid.
func isValid(t time.Time, timeout float64) bool {
	return time.Since(t).Minutes() < timeout
}

func init() {
	flag.BoolVar(&weatherHourly, "h", false, "Prints weather forecast hour by hour.")
	flag.BoolVar(&weatherWeek, "w", false, "Prints daily weather forecast for the next week.")
	flag.BoolVar(&airQuality, "a", false, "Prints air quality forecast.")
}

func runReports(f weather.Forecast, a []air.Forecast) {
	switch {
	case weatherHourly:
		report.WeatherHourly(f, a)
	case weatherWeek:
		report.WeatherWeek(f, a)
	case airQuality:
		report.AirQuality(f, a)
	default:
		report.Summary(f, a)
	}
}

func GetCoordinates() geolocation.Coordinates {
	// Get geolocation data.
	geoData := geolocation.GetGeoData(geolocation.IPAPIAddress)
	// Format coordinates and compose URLs for API calls.
	return geolocation.FormatCoordinates(geoData)
}

func RunReportsForFirstTime(c geolocation.Coordinates, t time.Time) (weather.Forecast, []air.Forecast) {
	dsURL := weather.BuildDarkSkyURL(weather.DarkSkyAddress, config.DarkSkyAPIKey, c, weather.DarkSkyUnits)
	// build AirNowURL
	anURL := air.BuildAirNowURL(air.AirNowAddress, c, t.Format("2006-01-02"), config.AirNowAPIKey)
	weatherChan := make(chan weather.Forecast)
	airChan := make(chan []air.Forecast)
	go func() {
		weatherChan <- weather.GetForecast(dsURL)
	}()
	go func() {
		airChan <- air.GetForecast(anURL)
	}()
	// Wait for API calls to return and run reports.
	weatherForecast = <-weatherChan
	close(weatherChan)
	airForecast = <-airChan
	close(airChan)
	runReports(weatherForecast, airForecast)
	report.TW.Flush()
	return weatherForecast, airForecast
}

func SaveForecasts(homeDir string, coordinates geolocation.Coordinates, weather weather.Forecast, air []air.Forecast) {
	// Update last api call
	storage.UpdateLastCall(coordinates, homeDir+storage.SavedCallFileName)

	// Save forecasts for next call
	storage.SaveWeatherForecast(homeDir+storage.SavedWeatherFileName, weatherForecast)
	storage.SaveAirForecast(homeDir+storage.SavedAirFileName, airForecast)
}

func main() {
	// Print time in order to signal program start and get date for building AirNow Url.
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05 MST 2006"))
	// Parse flags to determine which report to run.
	flag.Parse()

	// First get home directory for user.
	homeDir, err := storage.GetHomeDir()

	// If the home directory could not be determined, get coordinates
	// then call APIs, run reports, and exit.
	if err != nil {
		RunReportsForFirstTime(GetCoordinates(), t)
		// Since no home directory was found, skip caching forecast and exit.
		return
	}

	// Identify or create vaporwair directory.
	storage.CreateVaporwairDir(homeDir + storage.VaporwairDir)

	// Get Config
	cf := homeDir + storage.ConfigFileName
	config = storage.GetConfig(cf)

	// Get Coordinates from IP-API
	geoChan := make(chan geolocation.GeoData)
	go func() {
		geoChan <- geolocation.GetGeoData(geolocation.IPAPIAddress)
	}()

	// Channels to store calls with newly confirmed coordinates
	airChan := make(chan []air.Forecast)
	weatherChan := make(chan weather.Forecast)

	// Load previous call metadata to determine if call is still valid.
	pc, err := storage.LoadCallInfo(homeDir + storage.SavedCallFileName)
	if err != nil {
		// If not, run the reports for the first time.
		fmt.Println("No previous call info detected.")
		coordinates := GetCoordinates()
		weatherForecast, airForecast = RunReportsForFirstTime(coordinates, t)
		SaveForecasts(homeDir, coordinates, weatherForecast, airForecast)
		return
	}

	// If saved forecasts are found, check if the call has expired.
	valid := isValid(pc.Time, 5)

	// If the previous air and weather forecasts are still valid
	// i.e. they were made within the timeout period suppplied to the isValid function
	// (presumably from the same location), print forecast report and return
	if valid {
		// Load previous weather forecast from disk
		pwf, err := storage.LoadSavedWeather(homeDir + storage.SavedWeatherFileName)
		if err != nil {
			fmt.Println("No previous weather forecast found.")
		}
		// Load previous air forecast from disk
		paf, err := storage.LoadSavedAir(homeDir + storage.SavedAirFileName)
		if err != nil {
			fmt.Println("No previous air forecast found.")
			paf = <-airChan
		}

		runReports(pwf, paf)
		report.TW.Flush()
	}

	// Assume user has not changed coordinates since last weather check
	// and make optimistic call to APIs using saved coordinates.
	odsURL := weather.BuildDarkSkyURL(weather.DarkSkyAddress, config.DarkSkyAPIKey, pc.Coordinates, weather.DarkSkyUnits)
	// build AirNowURL
	oanURL := air.BuildAirNowURL(air.AirNowAddress, pc.Coordinates, t.Format("2006-01-02"), config.AirNowAPIKey)

	// optimistic channels
	ow := make(chan weather.Forecast)
	oa := make(chan []air.Forecast)

	go func() {
		ow <- weather.GetForecast(odsURL)
	}()
	go func() {
		oa <- air.GetForecast(oanURL)
	}()

	// Get geolocation data from channel and extract coordinates.
	geoData := <-geoChan
	coordinates := geolocation.FormatCoordinates(geoData)
	fmt.Println(coordinates.City, coordinates.Zip, "|", coordinates.Latitude, ",", coordinates.Longitude)

	// If coordinates returned by IP-API call differ from coordinates in saved forecast,
	// user is in a new location, and calls with the updated coordinates need to be made.

	// build DarkSkyURL
	dsURL := weather.BuildDarkSkyURL(weather.DarkSkyAddress, config.DarkSkyAPIKey, coordinates, weather.DarkSkyUnits)
	// build AirNowURL
	anURL := air.BuildAirNowURL(air.AirNowAddress, coordinates, t.Format("2006-01-02"), config.AirNowAPIKey)

	// Asynchronously make calls to Dark Sky and Airnow with confirmed coordinates
	go func() {
		weatherChan <- weather.GetForecast(dsURL)
	}()
	go func() {
		airChan <- air.GetForecast(anURL)
	}()

	weatherForecast = <-weatherChan
	close(weatherChan)
	airForecast = <-airChan
	close(airChan)
	runReports(weatherForecast, airForecast)
	report.TW.Flush()

	SaveForecasts(homeDir, coordinates, weatherForecast, airForecast)
}
