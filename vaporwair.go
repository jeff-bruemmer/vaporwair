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
	"os"
	"strings"
	"time"
)

// Flags
var weatherHourly bool
var weatherWeek bool
var airQuality bool
var clothingReport bool
var insightsReport bool
var zipCode string
var useCurrentLocation bool

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
	maxIterations := 600 // 60 seconds timeout (600 * 100ms)

	for i := 0; i <= maxIterations; i++ {
		select {
		case <-done:
			// Clear the line and exit
			fmt.Printf("\r                                                      ")
			result <- startTime
			return
		default:
			// Set the interval to add another =.
			time.Sleep(100 * time.Millisecond)

			// Update progress bar (loop it if necessary)
			if i < len(meterInit)-1 {
				meter = strings.Replace(meter, "> ", "=>", 1)
			}
			fmt.Print(meter)

			// Timeout after max iterations
			if i == maxIterations {
				fmt.Printf("\r                                                      ")
				log.Fatal("Request timed out after 60 seconds. The weather service may be unavailable.")
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
	case insightsReport:
		report.InsightsReport(f, a)
	default:
		report.InsightsReport(f, a)
	}
}

// GetCoordinates retrieves coordinates based on zip code flag, default zip, or IP address.
func GetCoordinates(appConfig storage.AppConfig) (geolocation.Coordinates, string) {
	var geoData geolocation.GeoData
	var err error
	var usedZip string

	// Priority: 1) -current flag (IP), 2) -zip flag, 3) default zip from config, 4) IP geolocation
	if useCurrentLocation {
		// Force IP-based location (temporarily override default zip)
		geoData, err = geolocation.GetGeoData(geolocation.IPAPIAddress)
		if err != nil {
			log.Fatalf("Failed to determine location from IP address: %v\nPlease check your internet connection.", err)
		}
		// usedZip remains empty so we don't update the default
	} else if zipCode != "" {
		// Get coordinates from zip code flag
		usedZip = zipCode
		geoData, err = geolocation.GetGeoDataFromZip(zipCode)
		if err != nil {
			log.Fatalf("Failed to get location for zip code %s: %v\nPlease verify the zip code is valid.", zipCode, err)
		}
	} else if appConfig.Config.DefaultZipCode != "" {
		// Get coordinates from saved default zip code
		usedZip = appConfig.Config.DefaultZipCode
		geoData, err = geolocation.GetGeoDataFromZip(appConfig.Config.DefaultZipCode)
		if err != nil {
			fmt.Printf("Warning: Could not use default zip code %s: %v\n", appConfig.Config.DefaultZipCode, err)
			fmt.Println("Falling back to IP-based location...")
			// Fall through to IP geolocation
			usedZip = ""
			geoData, err = geolocation.GetGeoData(geolocation.IPAPIAddress)
			if err != nil {
				log.Fatalf("Failed to determine location: %v\nPlease check your internet connection or specify a valid zip code with -zip flag.", err)
			}
		}
	} else {
		// Get geolocation data from IP address
		geoData, err = geolocation.GetGeoData(geolocation.IPAPIAddress)
		if err != nil {
			log.Fatalf("Failed to determine location from IP address: %v\nPlease check your internet connection or specify a zip code with -zip flag.", err)
		}
	}

	// Format coordinates and compose URLs for API calls.
	return geolocation.FormatCoordinates(geoData), usedZip
}

func PrintBanner() {
	banner := `
██╗   ██╗ █████╗ ██████╗  ██████╗ ██████╗ ██╗    ██╗ █████╗ ██╗██████╗
██║   ██║██╔══██╗██╔══██╗██╔═══██╗██╔══██╗██║    ██║██╔══██╗██║██╔══██╗
██║   ██║███████║██████╔╝██║   ██║██████╔╝██║ █╗ ██║███████║██║██████╔╝
╚██╗ ██╔╝██╔══██║██╔═══╝ ██║   ██║██╔══██╗██║███╗██║██╔══██║██║██╔══██╗
 ╚████╔╝ ██║  ██║██║     ╚██████╔╝██║  ██║╚███╔███╔╝██║  ██║██║██║  ██║
  ╚═══╝  ╚═╝  ╚═╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝
`
	fmt.Println(banner)
}

func PrintSpaceTime(t, t1 time.Time, c geolocation.Coordinates) {
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
	flag.Usage = func() {
		PrintBanner()
		fmt.Fprintf(os.Stderr, "\nUsage: vaporwair [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nWithout any flags, vaporwair displays the insights report (comparative analysis & weather recommendations).\n")
	}

	flag.BoolVar(&weatherHourly, "h", false, "Prints weather forecast hour by hour.")
	flag.BoolVar(&weatherWeek, "w", false, "Prints daily weather forecast for the next week.")
	flag.BoolVar(&airQuality, "a", false, "Prints air quality forecast.")
	flag.BoolVar(&clothingReport, "c", false, "Prints clothing recommendations based on weather (what to wair).")
	flag.BoolVar(&insightsReport, "i", false, "Prints comparative analysis and time-based insights.")
	flag.StringVar(&zipCode, "zip", "", "Get weather for a specific US zip code (e.g., -zip=10001).")
	flag.BoolVar(&useCurrentLocation, "current", false, "Use current IP-based location (temporary override).")
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
func fetchForecasts(coords geolocation.Coordinates, config storage.Config) (weather.Forecast, []air.Forecast, error) {
	weatherChan := make(chan weather.Forecast)
	airChan := make(chan []air.Forecast)
	errChan := make(chan error, 2)

	// Fetch weather forecast
	go func() {
		forecast, err := weather.GetNOAAWeatherForecast(coords)
		if err != nil {
			errChan <- fmt.Errorf("failed to retrieve weather forecast from NOAA: %w", err)
			return
		}
		weatherChan <- forecast
	}()

	// Fetch air quality forecast
	go func() {
		if config.AirNowAPIKey != "" && coords.Zip != "" {
			anURL := air.BuildAirNowURL(air.AirNowAddress, coords.Zip, config.AirNowAPIKey)
			forecast := air.GetForecast(anURL)
			airChan <- forecast
		} else {
			airChan <- []air.Forecast{}
		}
	}()

	// Wait for results with timeout
	select {
	case err := <-errChan:
		return weather.Forecast{}, nil, err
	case wf := <-weatherChan:
		af := <-airChan
		return wf, af, nil
	case <-time.After(30 * time.Second):
		return weather.Forecast{}, nil, fmt.Errorf("timeout: weather service did not respond within 30 seconds")
	}
}

// loadCachedForecasts implements optimistic cache retrieval.
// Returns true if cached forecasts were used, false if new forecasts are needed.
//
// Cache Invalidation Rules:
//   1. Time-based: Cache expires after CacheTimeoutMinutes (default: 5 minutes)
//   2. Location-based: Cache is tied to coordinates from last API call
//   3. Existence-based: Missing cache files trigger a fresh API call
//   4. Override flags: When -zip or -current flags are used, ignore cache
//   5. Default zip: When default zip code is set, ignore IP-based cache
//
// This optimistic approach prioritizes speed over freshness for recent queries.
func loadCachedForecasts(appConfig storage.AppConfig, t time.Time, spinnerDone chan bool, spinnerResult chan time.Time) bool {
	// Skip cache if user specified a zip code via flag (they want a specific location)
	if zipCode != "" {
		return false
	}

	// Skip cache if user requested current IP-based location
	if useCurrentLocation {
		return false
	}

	// Skip cache if default zip code is set (using zip-based location instead of IP)
	// This ensures we don't mix IP-based and zip-based forecasts
	if appConfig.Config.DefaultZipCode != "" {
		return false
	}

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
	PrintBanner()
	PrintSpaceTime(t, t1, pc.Coordinates)
	RunReports(pwf, paf)
	report.TW.Flush()

	return true
}

// main orchestrates the application flow: setup, caching, fetching, and reporting.
func main() {
	t := time.Now()
	flag.Parse()

	// Setup configuration BEFORE starting spinner (may prompt for user input)
	appConfig, err := setupConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	// Create channels for spinner coordination (no global state)
	spinnerDone := make(chan bool)
	spinnerResult := make(chan time.Time)

	// Start spinner in background
	go Spinner(t, spinnerDone, spinnerResult)

	// Try to use cached forecasts if still valid
	if loadCachedForecasts(appConfig, t, spinnerDone, spinnerResult) {
		return // Used cache successfully
	}

	// Cache miss or expired - need to fetch new forecasts
	coordinates, usedZip := GetCoordinates(appConfig)
	wf, af, err := fetchForecasts(coordinates, appConfig.Config)
	if err != nil {
		spinnerDone <- true
		<-spinnerResult
		log.Fatal(err)
	}

	// Stop spinner and display results
	spinnerDone <- true
	t1 := <-spinnerResult
	PrintBanner()
	PrintSpaceTime(t, t1, coordinates)
	RunReports(wf, af)
	report.TW.Flush()

	// Save forecasts for future use
	SaveForecasts(appConfig.HomeDir, coordinates, wf, af)

	// If a zip code was used (either from flag or default), save it as the new default
	if usedZip != "" && usedZip != appConfig.Config.DefaultZipCode {
		err = storage.UpdateDefaultZipCode(appConfig.HomeDir, usedZip)
		if err != nil {
			// Non-fatal error - just log it
			fmt.Printf("Note: Could not save default zip code: %v\n", err)
		}
	}
}
