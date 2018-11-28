package main

import (
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/dialer"
	"github.com/jeff-bruemmer/vaporwair/geolocation"
	"github.com/jeff-bruemmer/vaporwair/storage"
	"log"
	"os/user"
	"time"
)

// Has the forecast expired?
func expired(t time.Time, duration float64) bool {
	return time.Since(t).Minutes() > duration
}

// Dials IP-API to obtain geolocation data
func GetGeoData(address string) {
	resp, err := dialer.NetReq(dialer.IPAPIAddress, 5, false)
	if err != nil {
		fmt.Println("There was a problem obtaining your coordinates.")
	}
	var geoData geolocation.GeoData
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&geoData)
}

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

func main() {
	// Print time (also signals to user that program is running).
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05 MST 2006"))

	// Check for saved forecast.
	// First get home directory for user.
	homeDir, err := storage.GetHomeDir()
	if err != nil {
		log.Fatal("User could not be identified.\n", err)
	}
	fmt.Println(homeDir)

	// If saved forecast found, check if call has expired.

	// If still valid, print forecast report and return

	// Get Coordinates from IP-API

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

	usr, err := user.Current()
	if err != nil {
		log.Fatal("User could not be identified\n", err)
	}
	configFile := storage.ConfigFilePath(usr.HomeDir)
	fmt.Println("ConfigFile:", configFile)
}
