package air

import (
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"testing"
)

const exLatitude = "34.0308"
const exLongitude = "-118.473"
const exDate = "Date"
const exKey = "Key"
const exCity = "City"
const exZip = "Zip"

var exCoordinates = geolocation.Coordinates{exLatitude, exLongitude, exCity, exZip}

func TestBuildAirNowURL(t *testing.T) {
	got := BuildAirNowURL(AirNowAddress, exCoordinates, exDate, exKey)
	answer := AirNowAddress +
		"latitude=" + exCoordinates.Latitude +
		"&longitude=" + exCoordinates.Longitude +
		"&date=" + exDate +
		"&distance=25" +
		"&API_KEY=" + exKey
	if got != answer {
		t.Errorf("BuildAirNowURL(AirNowAddress, ex.Coordinates, exDate, exKey) = %s; want "+answer, got)
	}
}
