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
var weatherDaily bool
var weatherWeek bool
var weatherAlerts bool
var airQuality bool
var clothingReport bool
var insightsReport bool
var summaryReport bool
var zipCode string
var useCurrentLocation bool
var refresh bool

// isValid checks if cached forecast is still fresh (optimistic caching).
func isValid(t time.Time, timeout float64) bool {
	return time.Since(t).Minutes() < timeout
}

// Spinner creates a randomly populated loading bar with zero connection to reality.
// Its purpose is to show the user the program is running.
// It listens on the done channel and stops when signaled.
// Returns the original timestamp via the result channel.
func Spinner(startTime time.Time, done chan bool, result chan time.Time) {
	terminalWidth := report.GetTerminalWidth()
	barWidth := max(terminalWidth-2, 10) // Account for brackets []
	maxIterations := 600                 // 60 seconds timeout

	for i := 0; i <= maxIterations; i++ {
		select {
		case <-done:
			fmt.Printf("\r%s\r", strings.Repeat(" ", terminalWidth))
			result <- startTime
			return
		default:
			time.Sleep(100 * time.Millisecond)

			bar := make([]rune, barWidth)
			for j := range barWidth {
				if time.Now().UnixNano()%(int64(j+1)*3) == 0 {
					bar[j] = '█'
				} else {
					bar[j] = ' '
				}
			}

			fmt.Printf("\r[%s]", string(bar))

			if i == maxIterations {
				fmt.Printf("\r%s\r", strings.Repeat(" ", terminalWidth))
				log.Fatal("Request timed out after 60 seconds. The weather service may be unavailable.")
			}
		}
	}
}

// RunReports determines which report to run based on flags.
// Only one report can be run at a time.
func RunReports(f weather.Forecast, a []air.Forecast) {
	switch {
	case weatherHourly:
		report.WeatherHourly(f, a)
	case weatherDaily:
		report.WeatherDaily(f, a)
	case weatherWeek:
		report.WeatherWeek(f, a)
	case weatherAlerts:
		report.WeatherAlertsReport(f, a)
	case airQuality:
		report.AirQuality(f, a)
	case clothingReport:
		report.ClothingReport(f, a)
	case insightsReport:
		report.InsightsReport(f, a)
	case summaryReport:
		report.Summary(f, a)
	default:
		report.InsightsReport(f, a)
	}
}

// getIPGeoData retrieves geolocation data from IP address.
func getIPGeoData() geolocation.GeoData {
	geoData, err := geolocation.GetGeoData(geolocation.IPAPIAddress)
	if err != nil {
		log.Fatalf("Failed to determine location from IP address: %v\nPlease check your internet connection.", err)
	}
	return geoData
}

// GetCoordinates retrieves coordinates based on zip code flag, default zip, or IP address.
// Priority: 1) -current flag (IP), 2) -zip flag, 3) default zip from config, 4) IP geolocation
func GetCoordinates(appConfig storage.AppConfig) (geolocation.Coordinates, string) {
	// Force IP-based location (temporarily override default zip)
	if useCurrentLocation {
		return geolocation.FormatCoordinates(getIPGeoData()), ""
	}

	// Get coordinates from zip code flag
	if zipCode != "" {
		geoData, err := geolocation.GetGeoDataFromZip(zipCode)
		if err != nil {
			log.Fatalf("Failed to get location for zip code %s: %v\nPlease verify the zip code is valid.", zipCode, err)
		}
		return geolocation.FormatCoordinates(geoData), zipCode
	}

	// Try saved default zip code with fallback to IP
	if appConfig.Config.DefaultZipCode != "" {
		geoData, err := geolocation.GetGeoDataFromZip(appConfig.Config.DefaultZipCode)
		if err != nil {
			fmt.Printf("Warning: Could not use default zip code %s: %v\n", appConfig.Config.DefaultZipCode, err)
			fmt.Println("Falling back to IP-based location...")
			return geolocation.FormatCoordinates(getIPGeoData()), ""
		}
		return geolocation.FormatCoordinates(geoData), appConfig.Config.DefaultZipCode
	}

	// Default: use IP geolocation
	return geolocation.FormatCoordinates(getIPGeoData()), ""
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
	// Display location on two lines for better terminal width compatibility
	fmt.Printf("%s, %s\n", c.City, c.Zip)
	fmt.Printf("%s, %s\n", c.Latitude, c.Longitude)
}

// SaveForecasts persists forecasts to disk with timestamp and coordinates.
func SaveForecasts(homeDir string, coordinates geolocation.Coordinates, wf weather.Forecast, af []air.Forecast) {
	storage.UpdateLastCall(coordinates, homeDir+storage.SavedCallFileName)
	storage.SaveWeatherForecast(homeDir+storage.SavedWeatherFileName, wf)
	storage.SaveAirForecast(homeDir+storage.SavedAirFileName, af)
}
func init() {
	flag.Usage = func() {
		PrintBanner()
		fmt.Fprintf(os.Stderr, "\nUsage: vaporwair [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nWithout any flags, vaporwair displays the insights report (comparative analysis & weather recommendations).\n")
	}

	flag.BoolVar(&weatherHourly, "h", false, "Prints weather forecast hour by hour with detailed conditions.")
	flag.BoolVar(&weatherDaily, "d", false, "Prints comprehensive daily forecast with all NOAA fields.")
	flag.BoolVar(&weatherWeek, "w", false, "Prints daily weather forecast for the next week.")
	flag.BoolVar(&weatherAlerts, "alerts", false, "Prints active weather alerts and warnings.")
	flag.BoolVar(&airQuality, "a", false, "Prints air quality forecast.")
	flag.BoolVar(&clothingReport, "c", false, "Prints clothing recommendations based on weather (what to wair).")
	flag.BoolVar(&insightsReport, "i", false, "Prints comparative analysis and time-based insights.")
	flag.BoolVar(&summaryReport, "s", false, "Prints summary report with current conditions.")
	flag.StringVar(&zipCode, "zip", "", "Get weather for a specific US zip code (e.g., -zip=10001).")
	flag.BoolVar(&useCurrentLocation, "current", false, "Use current IP-based location (temporary override).")
	flag.BoolVar(&refresh, "refresh", false, "Force fresh API call, bypassing cache.")
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

// loadCachedForecasts loads forecasts from cache if valid.
// Cache is invalidated by: time expiry, location change, or user flags (-zip, -current, -refresh).
func loadCachedForecasts(appConfig storage.AppConfig, t time.Time, spinnerDone chan bool, spinnerResult chan time.Time) bool {
	if refresh || zipCode != "" || useCurrentLocation || appConfig.Config.DefaultZipCode != "" {
		return false
	}

	pc, err := storage.LoadCallInfo(appConfig.HomeDir + storage.SavedCallFileName)
	if err != nil || !isValid(pc.Time, appConfig.CacheTimeoutMinutes) {
		return false
	}
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
