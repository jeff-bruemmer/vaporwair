package geolocation

type Coordinates struct {
	Latitude  float64
	Longitude float64
	City      string
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

// Extracts Coordinates and City from GeoData struct.
// Coordinates will be saved to make optimistic API calls that assume user has not changed location.
func GetCoordinates(data GeoData) Coordinates {
	c := Coordinates{
		Latitude:  data.Lat,
		Longitude: data.Lon,
		City:      data.City,
	}
	return c
}
