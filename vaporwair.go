package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/air"
	"github.com/jeff-bruemmer/vaporwair/dialer"
	"github.com/jeff-bruemmer/vaporwair/geolocation"
	"github.com/jeff-bruemmer/vaporwair/storage"
	"github.com/jeff-bruemmer/vaporwair/weather"
	"github.com/jeff-bruemmer/vaporwair/report"
	"log"
	"os"
	"time"
)

// buildAirNowURL creates http address for dialer to call Dark Sky API.
func buildDarkSkyURL(addr string, apikey string, c geolocation.Coordinates, units string) string {
	return addr +
		apikey +
		"/" +
		c.Latitude +
		"," +
		c.Longitude +
		"?units=" +
		units
}

// buildAirNowURL creates http address for dialer to call Air Now API.
func buildAirNowURL(addr string, c geolocation.Coordinates, date string, apiKey string) string {
	return addr +
		"latitude=" + c.Latitude +
		"&longitude=" + c.Longitude +
		"&date=" + date +
		"&distance=25" +
		"&API_KEY=" + apiKey
}

// GetGeoData dials the IP-API server to obtain geolocation data
// based on user's IP address.
func GetGeoData(addr string) geolocation.GeoData {
	var gd geolocation.GeoData
	// Request coordinates from ip-api and specify timeout in seconds
	resp, err := dialer.NetReq(addr, 5, false)
	if err != nil {
		fmt.Println("The geolocation service could not resolve your coordinates.")
		os.Exit(1)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&gd)

	if gd.Status == "fail" {
		fmt.Println("The geolocation service could not resolve your coordinates.")
		os.Exit(1)
	}
	return gd
}

// GetWeatherForecast dials the Dark Sky API and returns a weather.Forecast.
func GetWeatherForecast(addr string) weather.Forecast {
	var wf weather.Forecast
	// Request coordinates from ip-api and specify timeout in seconds
	// Set gzip bool to true.
	resp, err := dialer.NetReq(addr, 5, true)
	if err != nil {
		log.Fatal(err)
	}
	// Unzip response
	defer resp.Body.Close()
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		fmt.Println("Error decoding gzip response from Dark Sky API.")
		log.Fatal(err)
	}
	// Decode unzipped response into weather forecast.
	defer gz.Close()
	json.NewDecoder(gz).Decode(&wf)
	return wf
}

// GetAirQualityForecast dials AirNow API and returns a slice of air.Forecast.
func GetAirQualityForecast(addr string) []air.Forecast {
	var af []air.Forecast
	resp, err := dialer.NetReq(addr, 10, false)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&af)
	return af
}

// GetWeatherAndAirForecasts fires 2 goroutines to get weather and air reports,
// and returns each in a channel.
func GetWeatherAndAirForecasts(dsURL, anURL string) (chan weather.Forecast, chan []air.Forecast) {
	weather := make(chan weather.Forecast)
	air := make(chan []air.Forecast)
	go func() {
		weather <- GetWeatherForecast(dsURL)
	}()
	go func() {
		air <- GetAirQualityForecast(anURL)
	}()
	return weather, air
}


// fastAirForecast returns the first valid forecast
// to return from an api call.
// oaf = old air forecast (using the old coordinates)
// naf = new air forecast (using the new coordinates)
func fastAirForecast(vc bool, nc geolocation.Coordinates, oaf, naf chan air.Forecast) air.Forecast {
	var af air.Forecast
	// If previously used coordinates are valid
	if vc {
		// Therefore either call is valid, so first call to return wins
		select {
		case af = <-oaf:
			break
		case af = <-naf:
			break
		}
		// Otherwise the old coordinates are invalid and only the new air forecast is valid.
	} else {
		af = <-naf
	}

	return af
}

// fastWeatherForecast returns the first valid forecast
// to return from an api call.
func fastWeatherForecast(vc bool, nc geolocation.Coordinates, owf, nwf chan weather.Forecast) weather.Forecast {
	var wf weather.Forecast
	// If previously used coordinates are valid
	if vc {
		// Either call is valid, first call to return wins
		select {
		case wf = <-owf:
			break
		case wf = <-nwf:
			break
		}
		// Otherwise the old coordinates are invalid and only the new air forecast is valid.
	} else {
		wf = <-nwf
	}

	return wf
}

// Has the forecast expired? Specify a duration (in minutes) that determines whether or not
// the forecast is still valid.
func expired(t time.Time, duration float64) bool {
	return time.Since(t).Minutes() < duration
}

func main() {
	// Get Coordinates from IP-API
	geoChan := make(chan geolocation.GeoData)
	go func () {
		geoChan <- GetGeoData(dialer.IPAPIAddress)
	}()
	// Print time in order to signal program start and get date for building AirNow Url.
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05 MST 2006"))

	// First get home directory for user.
	homeDir, err := storage.GetHomeDir()
	if err != nil {
		log.Fatal("Home directory could not be determined.\n", err)
	}

	// Identify or create vaporwair directory.
	storage.CreateVaporwairDir(homeDir + storage.VaporwairDir)

	// Load previous call metadata to determine if call is still valid.
	oc, err := storage.LoadCallInfo(homeDir + storage.SavedCallFileName)
	if err != nil {
		fmt.Println("No previous call info detected.")
	}

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

	// Get Config
	cf := homeDir + storage.ConfigFileName
	config := storage.GetConfig(cf)

	fmt.Println(oc.Time)
	// If saved forecasts are found, check if the call has expired.
	valid := expired(oc.Time, 1)


	// If the previous air and weather forecasts are still valid,
	// (i.e. they were made within the last x minutes, presumably from the same location)
	// print forecast report and return
	if (valid) {
		fmt.Println("Report from cache.")
		report.Today(owf, oaf)
		report.TW.Flush()
		return
	}	
	// Otherwise get updated forecast

	// Get geolocation data from channel and extract coordinates.
	geoData := <-geoChan
	coordinates := geolocation.FormatCoordinates(geoData)
	fmt.Println(coordinates.City, coordinates.Zip, "|", coordinates.Latitude, ",", coordinates.Longitude)

	// build DarkSkyURL
	dsURL := buildDarkSkyURL(dialer.DarkSkyAddress, config.DarkSkyAPIKey, coordinates, dialer.DarkSkyUnits)
	// build AirNowURL
	anURL := buildAirNowURL(dialer.AirNowAddress, coordinates, t.Format("2006-01-02"), config.AirNowAPIKey)
	// If cached forecast has expired, dial IP-API call

	// If saved coordinates exist, assume user is in same location and use coordinates to:

	// 1. Dial optimistic Dark Sky API

	// 2. Dial optimistic AirNow

	// If coordinates returned by IP-API call differ from coordinates in saved forecast,
	// user is in a new location, and calls with the updated coordinates need to be made.

	// Get weather and air forecasts.
	w, a := GetWeatherAndAirForecasts(dsURL, anURL)
	wf := <-w
	af := <-a
	report.Today(wf, af)
	report.TW.Flush()

	//Select fastest valid forecast that returns i.e. the first forecast that used the user's current coordinates.

        // Update last api call
	storage.UpdateLastCall(coordinates, homeDir + storage.SavedCallFileName)	

	// Save forecasts for next call
	storage.SaveWeatherForecast(homeDir + storage.SavedWeatherFileName, wf)
	storage.SaveAirForecast(homeDir + storage.SavedAirFileName, af)
}
