package main

import (
	"flag"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"github.com/jeff-bruemmer/vaporwair/src/report"
	"github.com/jeff-bruemmer/vaporwair/src/storage"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"log"
	"strings"
	"time"
)

// Flags
var weatherHourly bool
var weatherWeek bool
var airQuality bool
var clothingReport bool

// isValid checks if a cached forecast is still fresh based on elapsed time.
// This implements optimistic caching: we assume forecasts don't change frequently,
// so we serve cached data within the timeout window to reduce API calls and improve response time.
func isValid(t time.Time, timeout float64) bool {
	return time.Since(t).Minutes() < timeout
}

// Spinner creates a basic loading bar with zero connection to reality.
// Its purpose is to show the user the program is running.
// It listens on the done channel and stops when signaled.
// Returns the original timestamp via the result channel.
func Spinner(startTime time.Time, done chan bool, result chan time.Time) {
	meterInit := "\r[=>                                               ]"
	meter := meterInit
	for i := 0; i <= len(meterInit)-1; i++ {
		select {
		case <-done:
			// Clear the line and exit
			fmt.Printf("\r                                                      ")
			result <- startTime
			return
		default:
			// Set the interval to add another =.
			time.Sleep(100 * time.Millisecond)
			meter = strings.Replace(meter, "> ", "=>", 1)
			fmt.Print(meter)
			// Loop spinner if it completes before the reports have returned
			if i == len(meterInit)-1 {
				log.Fatal("There was a problem getting your forecast. Please check your internet connection.")
			}
		}
	}
}

// Prints the time it took to download (or retrieve from disk) the forecasts.
func PrintElapsedTime(t time.Time) {
	fmt.Printf("\rForecasts fetched in %v seconds.\n", time.Since(t).Seconds())
}

// RunReports determines which report to run based on flags.
// Only one report can be run at a time.
// Uses the presentation layer to abstract data access from report formatting.
func RunReports(f weather.Forecast, a []air.Forecast) {
	// Create view model for presentation layer
	// This separates data structures from report logic
	_ = report.NewForecastViewModel(f, a)

	// Note: Existing report functions still use raw data structures.
	// Future enhancement: Refactor reports to use view model methods instead.
	switch {
	case weatherHourly:
		report.WeatherHourly(f, a)
	case weatherWeek:
		report.WeatherWeek(f, a)
	case airQuality:
		report.AirQuality(f, a)
	case clothingReport:
		report.ClothingReport(f, a)
	default:
		report.Summary(f, a)
	}
}

// GetCoordinates retrieves user's current coordinates via IP address
// and the IP-API.
func GetCoordinates() geolocation.Coordinates {
	// Get geolocation data.
	geoData, err := geolocation.GetGeoData(geolocation.IPAPIAddress)
	if err != nil {
		fmt.Println("Error:", err)
		log.Fatal("Unable to determine your location. Please check your internet connection.")
	}
	// Format coordinates and compose URLs for API calls.
	return geolocation.FormatCoordinates(geoData)
}

func PrintSpaceTime(t, t1 time.Time, c geolocation.Coordinates) {
	PrintElapsedTime(t1)
	fmt.Println(t.Format("Mon Jan 2 15:04:05 MST 2006"))
	fmt.Println(c.City, c.Zip, "|", c.Latitude, ",", c.Longitude)
}

// SaveForecasts persists forecasts to disk for optimistic caching.
// Caching Strategy:
//   - Saves API response data along with timestamp and coordinates
//   - Subsequent requests within the timeout window will use cached data
//   - This reduces API load and provides faster response times for repeated queries
//   - Cache is location-aware: moving to a new location invalidates the cache
func SaveForecasts(homeDir string, coordinates geolocation.Coordinates, wf weather.Forecast, af []air.Forecast) {
	// Update last api call
	storage.UpdateLastCall(coordinates, homeDir+storage.SavedCallFileName)

	// Save forecasts for next call
	storage.SaveWeatherForecast(homeDir+storage.SavedWeatherFileName, wf)
	storage.SaveAirForecast(homeDir+storage.SavedAirFileName, af)
}

// Assign commandline flags.
func init() {
	flag.BoolVar(&weatherHourly, "h", false, "Prints weather forecast hour by hour.")
	flag.BoolVar(&weatherWeek, "w", false, "Prints daily weather forecast for the next week.")
	flag.BoolVar(&airQuality, "a", false, "Prints air quality forecast.")
	flag.BoolVar(&clothingReport, "c", false, "Prints clothing recommendations based on weather (what to wair).")
}

// setupConfiguration initializes the configuration directory and loads API keys.
// Returns the application configuration or an error if setup fails.
func setupConfiguration() (storage.AppConfig, error) {
	appConfig, err := storage.InitializeAppConfig()
	if err != nil {
		return appConfig, err
	}

	// Validate that AirNow API key is present
	if appConfig.Config.AirNowAPIKey == "" {
		configFile := appConfig.HomeDir + storage.ConfigFileName
		fmt.Println("Warning: AirNow API key is missing.")
		fmt.Println("Air quality data will not be available.")
		fmt.Println("To add your API key, edit: " + configFile)
		fmt.Println("Get a free API key at: https://docs.airnowapi.org/account/request/")
	}

	return appConfig, nil
}

// fetchForecasts retrieves weather and air quality forecasts for given coordinates.
// Returns weather forecast and air quality forecast.
func fetchForecasts(coords geolocation.Coordinates, config storage.Config, date string) (weather.Forecast, []air.Forecast, error) {
	weatherChan := make(chan weather.Forecast)
	airChan := make(chan []air.Forecast)
	errChan := make(chan error, 2)

	// Fetch weather forecast
	go func() {
		forecast, err := weather.GetNOAAWeatherForecast(coords)
		if err != nil {
			errChan <- fmt.Errorf("weather API error: %w", err)
			return
		}
		weatherChan <- forecast
	}()

	// Fetch air quality forecast
	go func() {
		if config.AirNowAPIKey != "" {
			anURL := air.BuildAirNowURL(air.AirNowAddress, coords, date, config.AirNowAPIKey)
			airChan <- air.GetForecast(anURL)
		} else {
			airChan <- []air.Forecast{}
		}
	}()

	// Wait for results
	select {
	case err := <-errChan:
		return weather.Forecast{}, nil, err
	case wf := <-weatherChan:
		af := <-airChan
		return wf, af, nil
	}
}

// loadCachedForecasts implements optimistic cache retrieval.
// Returns true if cached forecasts were used, false if new forecasts are needed.
//
// Cache Invalidation Rules:
//   1. Time-based: Cache expires after CacheTimeoutMinutes (default: 5 minutes)
//   2. Location-based: Cache is tied to coordinates from last API call
//   3. Existence-based: Missing cache files trigger a fresh API call
//
// This optimistic approach prioritizes speed over freshness for recent queries.
func loadCachedForecasts(appConfig storage.AppConfig, t time.Time, spinnerDone chan bool, spinnerResult chan time.Time) bool {
	pc, err := storage.LoadCallInfo(appConfig.HomeDir + storage.SavedCallFileName)
	if err != nil {
		return false // No cache available
	}

	// Check if cache is still valid
	if !isValid(pc.Time, appConfig.CacheTimeoutMinutes) {
		return false // Cache expired
	}

	// Load cached forecasts
	pwf, err := storage.LoadSavedWeather(appConfig.HomeDir + storage.SavedWeatherFileName)
	if err != nil {
		fmt.Println("No previous weather forecast found.")
		return false
	}

	paf, err := storage.LoadSavedAir(appConfig.HomeDir + storage.SavedAirFileName)
	if err != nil {
		fmt.Println("No previous air forecast found.")
		paf = []air.Forecast{}
	}

	// Stop spinner and print results
	spinnerDone <- true
	t1 := <-spinnerResult
	PrintSpaceTime(t, t1, pc.Coordinates)
	RunReports(pwf, paf)
	report.TW.Flush()

	return true
}

// main orchestrates the application flow: setup, caching, fetching, and reporting.
func main() {
	t := time.Now()
	flag.Parse()

	// Create channels for spinner coordination (no global state)
	spinnerDone := make(chan bool)
	spinnerResult := make(chan time.Time)

	// Start spinner in background
	go Spinner(t, spinnerDone, spinnerResult)

	// Setup configuration
	appConfig, err := setupConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	// Try to use cached forecasts if still valid
	if loadCachedForecasts(appConfig, t, spinnerDone, spinnerResult) {
		return // Used cache successfully
	}

	// Cache miss or expired - need to fetch new forecasts
	coordinates := GetCoordinates()
	wf, af, err := fetchForecasts(coordinates, appConfig.Config, t.Format("2006-01-02"))
	if err != nil {
		log.Fatal(err)
	}

	// Stop spinner and display results
	spinnerDone <- true
	t1 := <-spinnerResult
	PrintSpaceTime(t, t1, coordinates)
	RunReports(wf, af)
	report.TW.Flush()

	// Save forecasts for future use
	SaveForecasts(appConfig.HomeDir, coordinates, wf, af)
}
