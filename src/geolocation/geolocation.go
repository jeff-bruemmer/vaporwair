// This package handles data from ipwho.is requests, which uses IP addresses to
// obtain geolocation coordinates.
package geolocation

import (
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"strconv"
	"strings"
)

type Coordinates struct {
	Latitude  string
	Longitude string
	City      string
	Zip       string
}

type GeoData struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
	Query       string  `json:"query"`
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
	// Format coordinates for Forecast.io call
	c.Latitude = trimCoordinates(strconv.FormatFloat(gd.Lat, 'f', 10, 64))
	c.Longitude = trimCoordinates(strconv.FormatFloat(gd.Lon, 'f', 10, 64))
	c.City = gd.City
	c.Zip = gd.Zip
	return c
}

// ZipCodeResponse represents the response from zippopotam.us API
type ZipCodeResponse struct {
	PostCode    string `json:"post code"`
	Country     string `json:"country"`
	CountryAbbr string `json:"country abbreviation"`
	Places      []struct {
		PlaceName string `json:"place name"`
		Longitude string `json:"longitude"`
		State     string `json:"state"`
		StateAbbr string `json:"state abbreviation"`
		Latitude  string `json:"latitude"`
	} `json:"places"`
}

// ipWhoResponse represents the response from the ipwho.is API
type ipWhoResponse struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message"`
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	RegionCode  string  `json:"region_code"`
	City        string  `json:"city"`
	Postal      string  `json:"postal"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// GetGeoData dials the ipwho.is server to obtain geolocation data
// based on user's IP address.
func GetGeoData(addr string) (GeoData, error) {
	var gd GeoData
	// Request coordinates from ipwho.is and specify timeout in seconds
	resp, err := dialer.NetReq(addr, 5, false)
	if err != nil {
		return gd, fmt.Errorf("failed to connect to IP geolocation service (%s): %w", addr, err)
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

	// Convert to GeoData format
	gd.Status = "success"
	gd.Country = ipResp.Country
	gd.CountryCode = ipResp.CountryCode
	gd.Region = ipResp.RegionCode
	gd.RegionName = ipResp.Region
	gd.City = ipResp.City
	gd.Zip = ipResp.Postal
	gd.Lat = ipResp.Latitude
	gd.Lon = ipResp.Longitude
	gd.Query = ipResp.IP
	return gd, nil
}

// GetGeoDataFromZip retrieves geolocation data from a US zip code.
func GetGeoDataFromZip(zipCode string) (GeoData, error) {
	var gd GeoData

	// Validate zip code format
	if len(zipCode) != 5 {
		return gd, fmt.Errorf("invalid zip code format '%s' (must be 5 digits)", zipCode)
	}

	// Call zippopotam.us API
	url := ZipCodeAPIAddress + zipCode
	resp, err := dialer.NetReq(url, 5, false)
	if err != nil {
		return gd, fmt.Errorf("failed to connect to zip code lookup service: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode == 404 {
		return gd, fmt.Errorf("zip code '%s' not found - please verify it's a valid US zip code", zipCode)
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
		return gd, fmt.Errorf("no location data found for zip code '%s'", zipCode)
	}

	// Convert to GeoData format
	place := zipResp.Places[0]
	gd.Status = "success"
	gd.Country = zipResp.Country
	gd.CountryCode = zipResp.CountryAbbr
	gd.Region = place.StateAbbr
	gd.RegionName = place.State
	gd.City = place.PlaceName
	gd.Zip = zipResp.PostCode

	// Parse coordinates
	lat, err := strconv.ParseFloat(place.Latitude, 64)
	if err != nil {
		return gd, fmt.Errorf("invalid latitude data in zip code response: %w", err)
	}
	lon, err := strconv.ParseFloat(place.Longitude, 64)
	if err != nil {
		return gd, fmt.Errorf("invalid longitude data in zip code response: %w", err)
	}

	gd.Lat = lat
	gd.Lon = lon

	return gd, nil
}
