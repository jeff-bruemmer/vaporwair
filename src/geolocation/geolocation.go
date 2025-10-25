// This package handles data from IPAPI requests, which uses IP addresses to
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

const IPAPIAddress = "http://ip-api.com/json"
const ZipCodeAPIAddress = "https://api.zippopotam.us/us/"

// trimCoordinates drops trailing zeroes following
// conversion of coordinates from float64 to string
func trimCoordinates(c string) string {
	slice := strings.Split(c, "")
	for i := len(slice) - 1; i > 0; i-- {
		n := slice[i]
		if n == "0" {
			slice = slice[:i]
		} else {
			break
		}
	}
	return strings.Join(slice, "")
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

// GetGeoData dials the IP-API server to obtain geolocation data
// based on user's IP address.
func GetGeoData(addr string) (GeoData, error) {
	var gd GeoData
	// Request coordinates from ip-api and specify timeout in seconds
	resp, err := dialer.NetReq(addr, 5, false)
	if err != nil {
		return gd, fmt.Errorf("geolocation service error: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&gd)
	if err != nil {
		return gd, fmt.Errorf("error decoding geolocation response: %w", err)
	}

	if gd.Status == "fail" {
		return gd, fmt.Errorf("geolocation service could not resolve coordinates")
	}
	return gd, nil
}

// GetGeoDataFromZip retrieves geolocation data from a US zip code.
func GetGeoDataFromZip(zipCode string) (GeoData, error) {
	var gd GeoData

	// Call zippopotam.us API
	url := ZipCodeAPIAddress + zipCode
	resp, err := dialer.NetReq(url, 5, false)
	if err != nil {
		return gd, fmt.Errorf("zip code lookup error: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode == 404 {
		return gd, fmt.Errorf("zip code not found: %s", zipCode)
	}
	if resp.StatusCode != 200 {
		return gd, fmt.Errorf("zip code service returned status: %d", resp.StatusCode)
	}

	var zipResp ZipCodeResponse
	err = json.NewDecoder(resp.Body).Decode(&zipResp)
	if err != nil {
		return gd, fmt.Errorf("error decoding zip code response: %w", err)
	}

	if len(zipResp.Places) == 0 {
		return gd, fmt.Errorf("no location data found for zip code: %s", zipCode)
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
		return gd, fmt.Errorf("error parsing latitude: %w", err)
	}
	lon, err := strconv.ParseFloat(place.Longitude, 64)
	if err != nil {
		return gd, fmt.Errorf("error parsing longitude: %w", err)
	}

	gd.Lat = lat
	gd.Lon = lon

	return gd, nil
}
