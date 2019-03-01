package main

import (
	"flag"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/air"
	"github.com/jeff-bruemmer/vaporwair/geolocation"
	"github.com/jeff-bruemmer/vaporwair/storage"
	"github.com/jeff-bruemmer/vaporwair/weather"
	"github.com/jeff-bruemmer/vaporwair/dialer"
	"github.com/jeff-bruemmer/vaporwair/report"
	"log"
	"time"
)

// Is the forecast still valid? Specify a timeout duration (in minutes) that determines whether or not
// the forecast is still valid.
func isValid(t time.Time, timeout float64) bool {
	return time.Since(t).Minutes() < timeout
}

var weatherHourly bool
var weatherWeek bool
var airQuality bool

func init() {
	flag.BoolVar(&weatherHourly, "h", false, "Prints weather forecast hour by hour.")
	flag.BoolVar(&weatherWeek, "w", false, "Prints daily weather forecast for the next week.")
	flag.BoolVar(&airQuality, "a", false, "Prints air quality forecast.")
}

func runReports(f weather.Forecast, a []air.Forecast){
	switch {
	case weatherHourly:
		report.WeatherHourly(f, a)
	case weatherWeek:
		report.WeatherWeek(f,a)
	case airQuality:
		report.AirQuality(f, a)
	default:
		report.Summary(f, a)
	}
}

func main() {
	// Print time in order to signal program start and get date for building AirNow Url.
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05 MST 2006"))

	// First get home directory for user.
	homeDir, err := storage.GetHomeDir()
	if err != nil {
		log.Fatal("Home directory could not be determined.\n", err)
	}


	// Parse flags to determine which report to run.
	flag.Parse()

	// Get Coordinates from IP-API
	geoChan := make(chan geolocation.GeoData)
	go func () {
		geoChan <- geolocation.GetGeoData(dialer.IPAPIAddress)
	}()
	
	// Identify or create vaporwair directory.
	storage.CreateVaporwairDir(homeDir + storage.VaporwairDir)

	// Load previous call metadata to determine if call is still valid.
	oc, err := storage.LoadCallInfo(homeDir + storage.SavedCallFileName)
	if err != nil {
		fmt.Println("No previous call info detected.")
	}

   	// If saved forecasts are found, check if the call has expired.
	valid := isValid(oc.Time, 1)

	// If the previous air and weather forecasts are still valid,
	// (i.e. they were made within the last x minutes, presumably from the same location)
	// print forecast report and return
	if (valid) {
	// Load old weather forecast from disk
	owf, err := storage.LoadSavedWeather(homeDir + storage.SavedWeatherFileName)
	if err != nil {
		fmt.Println("No previous weather forecast found.")
	}

	// Load old air forecast from disk
	oaf, err := storage.LoadSavedAir(homeDir + storage.SavedAirFileName)
	if err != nil {
		fmt.Println("No previous air forecast found.")
	}

		report.Today(owf, oaf)
		report.TW.Flush()
		return
	}	

	// Otherwise get updated forecast

	// Get Config
	cf := homeDir + storage.ConfigFileName
	config := storage.GetConfig(cf)

	// Assume user has not changed coordinates since last weather check
	// and make optimistic call to APIs using saved coordinates.
	odsURL := weather.BuildDarkSkyURL(weather.DarkSkyAddress, config.DarkSkyAPIKey, oc.Coordinates, weather.DarkSkyUnits)
	// build AirNowURL
	oanURL := air.BuildAirNowURL(air.AirNowAddress, oc.Coordinates, t.Format("2006-01-02"), config.AirNowAPIKey)

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
	dsURL := weather.BuildDarkSkyURL(weather.DarkSkyAddress, config.DarkSkyAPIKey, coordinates, dialer.DarkSkyUnits)
	// build AirNowURL
	anURL := air.BuildAirNowURL(air.AirNowAddress, coordinates, t.Format("2006-01-02"), config.AirNowAPIKey)

	// Channels to store calls with newly confirmed coordinates
	airChan := make(chan []air.Forecast)
	weatherChan := make(chan weather.Forecast)

	// Asynchronously make calls to Dark Sky and Airnow with confirmed coordinates
	go func() {
		weatherChan <- weather.GetForecast(dsURL)
	}()
	go func() {
		airChan <- air.GetForecast(anURL)
	}()

	//Select fastest valid forecast that returns i.e. the first forecast that used the user's current coordinates.
	var air []air.Forecast
	var weather weather.Forecast

	weather = <- weatherChan
	close(weatherChan)
	air = <- airChan
	close(airChan)
	runReports(weather, air)
	report.TW.Flush()

        // Update last api call
	storage.UpdateLastCall(coordinates, homeDir + storage.SavedCallFileName)	

	// Save forecasts for next call
	storage.SaveWeatherForecast(homeDir + storage.SavedWeatherFileName, weather)
	storage.SaveAirForecast(homeDir + storage.SavedAirFileName, air)
}
