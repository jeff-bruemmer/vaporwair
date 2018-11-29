package geolocation

import (
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
