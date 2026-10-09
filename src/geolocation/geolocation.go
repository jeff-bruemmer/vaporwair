// This package handles data from ipwho.is requests, which uses IP addresses to
// obtain geolocation coordinates.
package geolocation

import (
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"strconv"
	"strings"
	"time"
)

type Coordinates struct {
	Latitude  string
	Longitude string
	City      string
	Zip       string
}

// GeoData is a location found from an IP address or a zip code.
type GeoData struct {
	City     string
	Zip      string
	Lat, Lon float64
}

// Use HTTPS to protect against MITM attacks that could leak location data
const IPAPIAddress = "https://ipwho.is/"
const ZipCodeAPIAddress = "https://api.zippopotam.us/us/"

// trimCoordinates drops trailing zeroes from coordinate strings
func trimCoordinates(c string) string {
	c = strings.TrimRight(c, "0")
	return strings.TrimRight(c, ".")
}

func FormatCoordinates(gd GeoData) Coordinates {
	var c Coordinates
	c.Latitude = trimCoordinates(strconv.FormatFloat(gd.Lat, 'f', 10, 64))
	c.Longitude = trimCoordinates(strconv.FormatFloat(gd.Lon, 'f', 10, 64))
	c.City = gd.City
	c.Zip = gd.Zip
	return c
}

// ZipCodeResponse represents the response from zippopotam.us API
type ZipCodeResponse struct {
	PostCode string `json:"post code"`
	Places   []struct {
		PlaceName string `json:"place name"`
		Longitude string `json:"longitude"`
		Latitude  string `json:"latitude"`
	} `json:"places"`
}

// ipWhoResponse represents the response from the ipwho.is API
type ipWhoResponse struct {
	Success   bool    `json:"success"`
	Message   string  `json:"message"`
	City      string  `json:"city"`
	Postal    string  `json:"postal"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// GetGeoData dials the ipwho.is server to obtain geolocation data
// based on user's IP address.
func GetGeoData(addr string) (GeoData, error) {
	var gd GeoData
	resp, err := dialer.Get(addr, 5*time.Second)
	if err != nil {
		return gd, err
	}
	defer resp.Body.Close()

	var ipResp ipWhoResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&ipResp)

	// Check HTTP status code, including the service's message when available
	if resp.StatusCode != 200 {
		if decodeErr == nil && ipResp.Message != "" {
			return gd, fmt.Errorf("IP geolocation service returned error status %d: %s", resp.StatusCode, ipResp.Message)
		}
		return gd, fmt.Errorf("IP geolocation service returned error status %d", resp.StatusCode)
	}

	if decodeErr != nil {
		return gd, fmt.Errorf("failed to parse geolocation response: %w", decodeErr)
	}

	if !ipResp.Success {
		return gd, fmt.Errorf("geolocation service could not determine location from your IP address: %s", ipResp.Message)
	}

	return GeoData{City: ipResp.City, Zip: ipResp.Postal, Lat: ipResp.Latitude, Lon: ipResp.Longitude}, nil
}

// ValidateZip reports whether zip is a 5-digit US zip code, with a hint for ZIP+4.
func ValidateZip(zip string) error {
	if len(zip) == 10 && zip[5] == '-' && isDigits(zip[:5]) && isDigits(zip[6:]) {
		return fmt.Errorf("zip code %q: use the 5-digit form, like %s", zip, zip[:5])
	}
	if len(zip) != 5 || !isDigits(zip) {
		return fmt.Errorf("zip code %q must be 5 digits, like 05401", zip)
	}
	return nil
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

// GetGeoDataFromZip retrieves geolocation data from a US zip code.
func GetGeoDataFromZip(zipCode string) (GeoData, error) {
	var gd GeoData

	if err := ValidateZip(zipCode); err != nil {
		return gd, err
	}

	resp, err := dialer.Get(ZipCodeAPIAddress+zipCode, 5*time.Second)
	if err != nil {
		return gd, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return gd, fmt.Errorf("zip code %s was not found; check that it is a valid US zip code", zipCode)
	}
	if resp.StatusCode != 200 {
		return gd, fmt.Errorf("zip code lookup service returned error status %d", resp.StatusCode)
	}

	var zipResp ZipCodeResponse
	err = json.NewDecoder(resp.Body).Decode(&zipResp)
	if err != nil {
		return gd, fmt.Errorf("failed to parse zip code response: %w", err)
	}

	if len(zipResp.Places) == 0 {
		return gd, fmt.Errorf("no location data found for zip code %s", zipCode)
	}

	place := zipResp.Places[0]
	lat, err := strconv.ParseFloat(place.Latitude, 64)
	if err != nil {
		return gd, fmt.Errorf("invalid latitude data in zip code response: %w", err)
	}
	lon, err := strconv.ParseFloat(place.Longitude, 64)
	if err != nil {
		return gd, fmt.Errorf("invalid longitude data in zip code response: %w", err)
	}

	return GeoData{City: place.PlaceName, Zip: zipResp.PostCode, Lat: lat, Lon: lon}, nil
}
