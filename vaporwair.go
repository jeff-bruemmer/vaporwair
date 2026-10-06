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

// notes collects non-fatal warnings raised while fetching.
// They're printed after the report so they don't garble the spinner line.
// Only the main goroutine appends to it.
var notes []string

// stopSpinner clears the spinner if one is running. fatalf calls it so an error
// message never lands on the same line as the spinner.
var stopSpinner = func() {}

// fatalf stops the spinner, prints an error to stderr, and exits.
func fatalf(format string, args ...any) {
	stopSpinner()
	log.Fatalf(format, args...)
}

// isValid checks if cached forecast is still fresh (optimistic caching).
func isValid(t time.Time, timeout float64) bool {
	return time.Since(t).Minutes() < timeout
}

// startSpinner shows an ASCII spinner and label on stderr until stopSpinner is called.
// It draws nothing when stderr isn't a terminal, so piped or redirected output stays clean.
func startSpinner(label string) {
	if !report.IsTerminal(os.Stderr) {
		return
	}
	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		frames := `|/-\`
		tick := time.NewTicker(100 * time.Millisecond)
		defer tick.Stop()
		for i := 0; ; i++ {
			fmt.Fprintf(os.Stderr, "\r%c %s", frames[i%len(frames)], label)
			select {
			case <-done:
				fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", len(label)+2))
				return
			case <-tick.C:
			}
		}
	}()
	stopSpinner = func() {
		close(done)
		<-finished
		stopSpinner = func() {}
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
		fatalf("Failed to determine location from IP address: %v\nTry specifying a location with -zip <code>.", err)
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
			fatalf("Failed to get location for zip code %s: %v\nPlease verify the zip code is valid.", zipCode, err)
		}
		return geolocation.FormatCoordinates(geoData), zipCode
	}

	// Try saved default zip code with fallback to IP
	if appConfig.Config.DefaultZipCode != "" {
		geoData, err := geolocation.GetGeoDataFromZip(appConfig.Config.DefaultZipCode)
		if err != nil {
			notes = append(notes, fmt.Sprintf("Could not use default zip code %s (%v); used IP-based location", appConfig.Config.DefaultZipCode, err))
			return geolocation.FormatCoordinates(getIPGeoData()), ""
		}
		return geolocation.FormatCoordinates(geoData), appConfig.Config.DefaultZipCode
	}

	// Default: use IP geolocation
	return geolocation.FormatCoordinates(getIPGeoData()), ""
}

func PrintBanner() {
	banner := `
 _   __ ___    ____   ____   ____  _      __ ___    ____ ____
| | / //   |  / __ \ / __ \ / __ \| | /| / //   |  /  _// __ \
| |/ // /| | / /_/ // / / // /_/ /| |/ |/ // /| |  / / / /_/ /
|   // ___ |/ ____// /_/ // _, _/ |  /|  // ___ |_/ / / _, _/
|__//_/  |_/_/     \____//_/ |_|  |__/|__//_/  |_/___//_/ |_|
`
	fmt.Fprintln(os.Stderr, banner)
}

// PrintHeader prints a one-line header: location, time, and cache age when served from cache.
// A zero cachedAt means the forecast was just fetched.
func PrintHeader(t time.Time, c geolocation.Coordinates, cachedAt time.Time) {
	parts := []string{}
	if place := strings.TrimSpace(c.City + " " + c.Zip); place != "" {
		parts = append(parts, report.Bold(place))
	}
	parts = append(parts, t.Format("Mon Jan 2, 15:04 MST"))
	if !cachedAt.IsZero() {
		age := t.Sub(cachedAt)
		if age < time.Minute {
			parts = append(parts, "cached just now")
		} else {
			parts = append(parts, fmt.Sprintf("cached %dm ago", int(age.Minutes())))
		}
	}
	fmt.Println(strings.Join(parts, " | "))
	fmt.Println()
}

// PrintNotes prints any warnings collected while fetching, after the report.
func PrintNotes() {
	if len(notes) > 0 {
		fmt.Println()
	}
	for _, n := range notes {
		fmt.Println("Note: " + n)
	}
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
		fmt.Fprintf(os.Stderr, "\nWithout any flags, vaporwair displays the insights report (today's forecast, what to wear, next few hours).\n")
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

	// Explain a missing AirNow key once, at setup. Later runs get a one-line note
	// only on reports that show air quality.
	if appConfig.Config.AirNowAPIKey == "" {
		configFile := appConfig.HomeDir + storage.ConfigFileName
		if appConfig.FirstRun {
			fmt.Println("No AirNow API key: air quality data will not be available.")
			fmt.Println("Get a free key at https://docs.airnowapi.org/account/request/")
			fmt.Println("and add it to " + configFile)
			fmt.Println()
		} else if airQuality || summaryReport {
			notes = append(notes, "Air quality off: no AirNow API key in "+configFile)
		}
	}

	return appConfig, nil
}

// fetchForecasts retrieves weather and air quality forecasts for given coordinates.
// Returns weather forecast and air quality forecast.
func fetchForecasts(coords geolocation.Coordinates, config storage.Config) (weather.Forecast, []air.Forecast, error) {
	type airResult struct {
		forecast []air.Forecast
		note     string // non-empty when AirNow failed; reported, not fatal
	}
	weatherChan := make(chan weather.Forecast)
	airChan := make(chan airResult, 1)
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
			forecast, err := air.GetForecast(anURL)
			var note string
			if err != nil {
				note = fmt.Sprintf("Air quality unavailable (AirNow: %v)", err)
			}
			airChan <- airResult{forecast, note}
		} else {
			airChan <- airResult{forecast: []air.Forecast{}}
		}
	}()

	// Wait for results with timeout
	select {
	case err := <-errChan:
		return weather.Forecast{}, nil, err
	case wf := <-weatherChan:
		ar := <-airChan
		if ar.note != "" {
			notes = append(notes, ar.note)
		}
		return wf, ar.forecast, nil
	case <-time.After(30 * time.Second):
		return weather.Forecast{}, nil, fmt.Errorf("timeout: weather service did not respond within 30 seconds")
	}
}

// targetZip returns the zip code this run asks for: the -zip flag, else the saved default.
// Empty means IP-based location (no default saved, or -current).
func targetZip(appConfig storage.AppConfig) string {
	if useCurrentLocation {
		return ""
	}
	if zipCode != "" {
		return zipCode
	}
	return appConfig.Config.DefaultZipCode
}

// loadCachedForecasts returns forecasts from the cache when they are fresh and for the
// location this run asks for. -refresh and -current always bypass the cache; a zip
// request only uses a cache saved for that same zip.
func loadCachedForecasts(appConfig storage.AppConfig) (storage.APICallInfo, weather.Forecast, []air.Forecast, bool) {
	var none storage.APICallInfo
	if refresh || useCurrentLocation {
		return none, weather.Forecast{}, nil, false
	}

	pc, err := storage.LoadCallInfo(appConfig.HomeDir + storage.SavedCallFileName)
	if err != nil || !isValid(pc.Time, appConfig.CacheTimeoutMinutes) {
		return none, weather.Forecast{}, nil, false
	}
	if zip := targetZip(appConfig); zip != "" && pc.Coordinates.Zip != zip {
		return none, weather.Forecast{}, nil, false
	}

	pwf, err := storage.LoadSavedWeather(appConfig.HomeDir + storage.SavedWeatherFileName)
	if err != nil {
		return none, weather.Forecast{}, nil, false
	}

	paf, err := storage.LoadSavedAir(appConfig.HomeDir + storage.SavedAirFileName)
	if err != nil {
		notes = append(notes, "Air quality unavailable (no cached air forecast)")
		paf = []air.Forecast{}
	}

	return pc, pwf, paf, true
}

// render prints the header, the selected report, and any notes.
// A zero cachedAt means the forecast was just fetched.
func render(t time.Time, c geolocation.Coordinates, cachedAt time.Time, wf weather.Forecast, af []air.Forecast) {
	PrintHeader(t, c, cachedAt)
	RunReports(wf, af)
	report.TW.Flush()
	PrintNotes()
}

// saveDefaultZip remembers a -zip request as the default location for later runs.
func saveDefaultZip(appConfig storage.AppConfig) {
	if zipCode == "" || zipCode == appConfig.Config.DefaultZipCode {
		return
	}
	if err := storage.UpdateDefaultZipCode(appConfig.HomeDir, zipCode); err != nil {
		fmt.Printf("Note: Could not save default zip code: %v\n", err)
	}
}

// main orchestrates the application flow: setup, caching, fetching, and reporting.
func main() {
	t := time.Now()
	log.SetFlags(0) // errors read as CLI messages, not timestamped log lines
	flag.Parse()

	// Setup configuration first (may prompt for user input)
	appConfig, err := setupConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	// Serve from cache when it is fresh and for the requested location
	if pc, wf, af, ok := loadCachedForecasts(appConfig); ok {
		render(t, pc.Coordinates, pc.Time, wf, af)
		saveDefaultZip(appConfig)
		return
	}

	// Cache miss or expired - fetch new forecasts
	startSpinner("Fetching forecast...")
	coordinates, _ := GetCoordinates(appConfig)
	wf, af, err := fetchForecasts(coordinates, appConfig.Config)
	if err != nil {
		fatalf("%v", err)
	}
	stopSpinner()

	render(t, coordinates, time.Time{}, wf, af)
	SaveForecasts(appConfig.HomeDir, coordinates, wf, af)
	saveDefaultZip(appConfig)
}
