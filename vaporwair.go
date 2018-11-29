package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/dialer"
	"github.com/jeff-bruemmer/vaporwair/geolocation"
	"github.com/jeff-bruemmer/vaporwair/storage"
	"github.com/jeff-bruemmer/vaporwair/weather"
	"log"
	"os"
	"time"
)

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

// GetWeatherForecast dials the Dark Sky API and returns a weather forecast.
func GetWeatherForecast(addr string) (weather.Forecast, error) {
	var wf weather.Forecast
	// Request coordinates from ip-api and specify timeout in seconds
	// Set gzip bool to true.
	resp, err := dialer.NetReq(addr, 5, true)
	if err != nil {
		return wf, err
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
	return wf, err
}

// printTime prints time and signals to user that program has started.
func printTime() {
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05 MST 2006"))
}

func main() {
	printTime()
	// First get home directory for user.
	homeDir, err := storage.GetHomeDir()
	if err != nil {
		log.Fatal("User could not be identified.\n", err)
	}
	// Check for saved forecast.

	// If saved forecast found, check if call has expired.

	// Get Config
	cf := storage.ConfigFilePath(homeDir, storage.ConfigFileName)
	config := storage.GetConfig(cf)
	fmt.Println(config)
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
	fmt.Println(dsURL)
	// build AirNowURL

	// If cached forecast has expired, dial IP-API call
	// If saved coordinates exist, assume user is in same location and use coordinates to:
	// 1. Dial optimistic Dark Sky API

	// 2. Dial optimistic AirNow

	// If coordinates returned by IP-API call differ from coordinates in saved forecast,
	// user is in a new location, and calls with the updated coordinates need to be made.
	// 1. Dial Dark Sky API
	wf, err := GetWeatherForecast(dsURL)
	if err != nil {
		log.Fatal("There was a problem obtaining the weather forecast.")
	}
	fmt.Println(wf)
	// 2. Dial AirNow API

	// Select fastest valid forecast that returns
	// i.e. the first forecast that used the user's current coordinates.

	// Print report

	// Save forecast for next call
}
