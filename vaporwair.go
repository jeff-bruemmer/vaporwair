package main

import (
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/dialer"
	"github.com/jeff-bruemmer/vaporwair/geolocation"
	"github.com/jeff-bruemmer/vaporwair/storage"
	"log"
	"os"
	"time"
)

// // Dials IP-API to obtain geolocation data
// func GetGeoData(address string) {
// 	resp, err := dialer.NetReq(dialer.IPAPIAddress, 5, false)
// 	if err != nil {
// 		fmt.Println("There was a problem obtaining your coordinates.")
// 	}
// 	var geoData geolocation.GeoData
// 	defer resp.Body.Close()
// 	json.NewDecoder(resp.Body).Decode(&geoData)
// }

func buildDarkSkyURL(addr string, apikey string, c geolocation.Coordinates, units string) string {
	return addr +
		apikey +
		"/" +
		fmt.Sprintf("%f", c.Latitude) +
		"," +
		fmt.Sprintf("%f", c.Longitude) +
		"?units=" +
		units
}

// GetGeoData dials the IP-API server to obtain geolocation data
// based on user's IP address.
func GetGeoData() (geolocation.GeoData, error) {
	var gd geolocation.GeoData
	// Request coordinates from ip-api and specify timeout in seconds
	resp, err := dialer.NetReq("http://ip-api.com/json", 2, false)
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

func main() {
	// Print time (also signals to user that program is running).
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05 MST 2006"))

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
	geoData, err = GetGeoData()
	if err != nil {
		fmt.Println("There was a problem obtaining your coordinates")
		log.Fatal(err)
	}
	fmt.Println(geoData)

	// build DarkSkyURL

	// build AirNowURL

	// If cached forecast has expired, dial IP-API call
	// If saved coordinates exist, assume user is in same location and use coordinates to:
	// 1. Dial optimistic Dark Sky API

	// 2. Dial optimistic AirNow

	// If coordinates returned by IP-API call differ from coordinates in saved forecast,
	// user is in a new location, and calls with the updated coordinates need to be made.
	// 1. Dial Dark Sky API
	// 2. Dial AirNow API

	// Select fastest valid forecast that returns
	// i.e. the first forecast that used the user's current coordinates.

	// Print report

	// Save forecast for next call
}
