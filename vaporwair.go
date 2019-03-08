package main

import (
	"flag"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/air"
	"github.com/jeff-bruemmer/vaporwair/geolocation"
	"github.com/jeff-bruemmer/vaporwair/report"
	"github.com/jeff-bruemmer/vaporwair/storage"
	"github.com/jeff-bruemmer/vaporwair/weather"
	"log"
	"strings"
	"time"
)

// Timeout, an int representing minutes, determines how long a forecast is valid.
const Timeout = 5

// Flags
var weatherHourly bool
var weatherWeek bool
var airQuality bool

// Globals
var weatherForecast weather.Forecast
var airForecast []air.Forecast
var config storage.Config

// Variables used to sync spinner.
var reportsReady = false
var spinnerChan = make(chan time.Time)

// Is the forecast still valid? Specify a timeout duration (in minutes) that determines whether or not
// the forecast is still valid.
func isValid(t time.Time, timeout float64) bool {
	return time.Since(t).Minutes() < timeout
}

// Spinner creates a basic loading bar with zero connection to reality.
// Its purpose is to show the user the program is running.
// It returns a time once it finds the the reports are ready to print.
// That time is put in a channel used to terminate the spinner.
func Spinner(t time.Time) time.Time {
	meterInit := "\r[=>                                               ]"
	meter := meterInit
	for i := 0; i <= len(meterInit)-1; i++ {
		if reportsReady {
			break
		}
		// Set the interval to add another =.
		time.Sleep(100 * time.Millisecond)
		meter = strings.Replace(meter, "> ", "=>", 1)
		fmt.Printf(meter)
		// Loop spinner if it completes before the reports have returned
		if i == len(meterInit)-1 {
			log.Fatal("There was a problem getting your forecast. Please check your internet connection.")
		}
	}
	// Clear the line.
	fmt.Printf("\r                                                      ")
	return t
}

// Prints the time it took to download (or retrieve from disk) the forecasts.
func PrintElapsedTime(t time.Time) {
	fmt.Printf("\rForecasts fetched in %v seconds.\n", time.Since(t).Seconds())
}

// RunReports determines which report to run based on flags.
// Only one report can be run at a time.
func RunReports(f weather.Forecast, a []air.Forecast) {
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

// GetCoordinates retrieves user's current coordinates via IP address
// and the IP-API.
func GetCoordinates() geolocation.Coordinates {
	// Get geolocation data.
	geoData := geolocation.GetGeoData(geolocation.IPAPIAddress)
	// Format coordinates and compose URLs for API calls.
	return geolocation.FormatCoordinates(geoData)
}

func PrintSpaceTime(t, t1 time.Time, c geolocation.Coordinates) {
	PrintElapsedTime(t1)
	fmt.Println(t.Format("Mon Jan 2 15:04:05 MST 2006"))
	fmt.Println(c.City, c.Zip, "|", c.Latitude, ",", c.Longitude)
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
	reportsReady = true
	t1 := <-spinnerChan
	close(spinnerChan)
	PrintSpaceTime(t, t1, c)
	RunReports(weatherForecast, airForecast)
	report.TW.Flush()
	return weatherForecast, airForecast
}

// Saves forecast in Vaporwair home directory for caching.
func SaveForecasts(homeDir string, coordinates geolocation.Coordinates, weather weather.Forecast, air []air.Forecast) {
	// Update last api call
	storage.UpdateLastCall(coordinates, homeDir+storage.SavedCallFileName)

	// Save forecasts for next call
	storage.SaveWeatherForecast(homeDir+storage.SavedWeatherFileName, weatherForecast)
	storage.SaveAirForecast(homeDir+storage.SavedAirFileName, airForecast)
}

// CaptureAPIKeys prompts users for Dark Sky and Air Now API keys
// and saves thems in a config file.
func CaptureAPIKeys(homeDir string) {
	DSAPIKey := storage.Capture("Enter Dark Sky API key: ")
	ANAPIKey := storage.Capture("Enter Air Now API key: ")
	err := storage.CreateConfig(homeDir, DSAPIKey, ANAPIKey)
	if err != nil {
		log.Fatal("There was a problem saving your APIkeys.")
	}
}

// CleanPrintSave closes open air, weather, and spinner channels; prints reports;
// and saves the forecasts for future reporting.
func CleanPrintSave(weatherChan chan weather.Forecast, airChan chan []air.Forecast, t time.Time, c geolocation.Coordinates, homeDir string) {
	// Wait for forecasts, then close channels.
	weatherForecast = <-weatherChan
	close(weatherChan)
	airForecast = <-airChan
	close(airChan)

	// Stop Spinner
	reportsReady = true
	t1 := <-spinnerChan
	close(spinnerChan)

	// Print time and geodata, then print reports.
	PrintSpaceTime(t, t1, c)
	RunReports(weatherForecast, airForecast)
	report.TW.Flush()

	// Save forecasts
	SaveForecasts(homeDir, c, weatherForecast, airForecast)
	return

}

// Assign commandline flags.
func init() {
	flag.BoolVar(&weatherHourly, "h", false, "Prints weather forecast hour by hour.")
	flag.BoolVar(&weatherWeek, "w", false, "Prints daily weather forecast for the next week.")
	flag.BoolVar(&airQuality, "a", false, "Prints air quality forecast.")
}

// The main function is large for a Go program, but it provides a good
// overview of the program's execution.
func main() {
	t := time.Now()
	// Start call to IP-API in case previously used coordinates either
	// do not exist or are invalid.
	geoChan := make(chan geolocation.GeoData)
	go func() {
		geoChan <- geolocation.GetGeoData(geolocation.IPAPIAddress)
	}()

	// Start Spinner
	go func() {
		spinnerChan <- Spinner(t)
	}()

	// Parse flags to determine which report to run.
	flag.Parse()

	// First get home directory for user.
	homeDir, err := storage.GetHomeDir()
	// If the home directory could not be determined, bail.
	if err != nil {
		log.Fatal("Unable to determine home directory.")
	}

	// Identify or create vaporwair directory.
	storage.CreateVaporwairDir(homeDir + storage.VaporwairDir)

	// Check if configuration file with API keys exists.
	cf := homeDir + storage.ConfigFileName
	configExists, _ := storage.Exists(cf)

	// If not, prompt user for API keys and create configuration file.
	if !configExists {
		CaptureAPIKeys(homeDir)
	}

	// Load API keys
	config = storage.GetConfig(cf)

	// Channels to store calls with newly confirmed coordinates
	airChan := make(chan []air.Forecast)
	weatherChan := make(chan weather.Forecast)

	// Load previous call metadata to determine if call is still valid.
	pc, err := storage.LoadCallInfo(homeDir + storage.SavedCallFileName)
	if err != nil {
		// If not, run the reports for the first time.
		coordinates := GetCoordinates()
		weatherForecast, airForecast = RunReportsForFirstTime(coordinates, t)
		SaveForecasts(homeDir, coordinates, weatherForecast, airForecast)
		return
	}

	// If saved forecasts are found, check if the call has expired.
	valid := isValid(pc.Time, Timeout)

	// If the previous air and weather forecasts are still valid
	// i.e. they were made within the timeout period suppplied to the isValid function
	// (presumably from the same location), print forecast report and return.
	if valid {
		// Load previous weather forecast from disk.
		pwf, err := storage.LoadSavedWeather(homeDir + storage.SavedWeatherFileName)
		if err != nil {
			fmt.Println("No previous weather forecast found.")
		}
		// Load previous air forecast from disk.
		paf, err := storage.LoadSavedAir(homeDir + storage.SavedAirFileName)
		if err != nil {
			fmt.Println("No previous air forecast found.")
			paf = <-airChan
		}
		reportsReady = true
		t1 := <-spinnerChan
		PrintSpaceTime(t, t1, pc.Coordinates)
		RunReports(pwf, paf)
		report.TW.Flush()
		return
	}

	// While waiting for the coordinates to return form the IP-API,
	// assume user has not changed coordinates since last weather check
	// and make optimistic call to APIs using saved coordinates.
	odsURL := weather.BuildDarkSkyURL(weather.DarkSkyAddress, config.DarkSkyAPIKey, pc.Coordinates, weather.DarkSkyUnits)
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

	// If current coordinates match previous coordinates, the optimistic API calls
	// are valid, no need to make new calls. Clean up, print reports, save forecasts,
	// and return.
	if coordinates.Latitude == pc.Coordinates.Latitude &&
		coordinates.Longitude == pc.Coordinates.Longitude {
		CleanPrintSave(ow, oa, t, coordinates, homeDir)
		return
	} else {
		// If coordinates returned by IP-API call differ from coordinates in saved forecast,
		// user is in a new location, and calls with the updated coordinates need to be made.

		// Build URLs.
		dsURL := weather.BuildDarkSkyURL(weather.DarkSkyAddress, config.DarkSkyAPIKey, coordinates, weather.DarkSkyUnits)
		anURL := air.BuildAirNowURL(air.AirNowAddress, coordinates, t.Format("2006-01-02"), config.AirNowAPIKey)

		// Asynchronously make calls to Dark Sky and Airnow with confirmed coordinates
		go func() {
			weatherChan <- weather.GetForecast(dsURL)
		}()
		go func() {
			airChan <- air.GetForecast(anURL)
		}()

		// Wait for forecasts to return, then clean up, print reports, save forecasts, and return.
		CleanPrintSave(weatherChan, airChan, t, coordinates, homeDir)
		return
	}
}
