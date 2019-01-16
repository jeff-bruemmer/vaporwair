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
func GetGeoData(addr string) (geolocation.GeoData, error) {
	var gd geolocation.GeoData
	// Request coordinates from ip-api and specify timeout in seconds
	resp, err := dialer.NetReq(addr, 5, false)
	if err != nil {
		return gd, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&gd)

	if gd.Status == "fail" {
		fmt.Println("The geolocation service could not resolve your coordinates.")
		os.Exit(1)
	}
	return gd, err

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
// and returns each in a channel. If either call fails, the program exits.
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

func main() {
	// Print time to signal program start and get date for building AirNow Url.
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05 MST 2006"))

	// First get home directory for user.
	homeDir, err := storage.GetHomeDir()
	if err != nil {
		log.Fatal("Home directory could not be determined.\n", err)
	}
	// Identify or create vaporwair directory.
	storage.CreateVaporwairDir(homeDir + "/.vaporwair")
	// TODO Check for saved forecast.

	// TODO If saved forecast found, check if call has expired.

	// Get Config
	cf := storage.FilePath(homeDir, storage.ConfigFileName)
	config := storage.GetConfig(cf)
	// If still valid, print forecast report and return

	// Get Coordinates from IP-API
	var geoData geolocation.GeoData
	geoData, err = GetGeoData(dialer.IPAPIAddress)
	if err != nil {
		fmt.Println("There was a problem obtaining your coordinates.")
		log.Fatal(err)
	}
	coordinates := geolocation.FormatCoordinates(geoData)
	fmt.Println(coordinates)
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
	wr := <-w
	ar := <-a
	report.CurrentTemp(wr)
	report.MinTemp(wr)
	report.MaxTemp(wr)
	report.AirQualityIndex(ar)
	report.TW.Flush()
	//Select fastest valid forecast that returns i.e. the first forecast that used the user's current coordinates.

        // Update last api call
	storage.UpdateLastCall(coordinates, homeDir + storage.SavedCallFileName)	
	// TODO Save weather forecast for next call
	storage.SaveWeatherForecast(storage.FilePath(homeDir, storage.SavedWeatherFileName), wr)


}
