package air

type Category struct {
	Number int    `json:"Number"`
	Name   string `json:"Name"`
}

type Forecast struct {
	DateIssue     string      `json:"DateIssue"`
	DateForecast  string      `json:"DateForecast"`
	ReportingArea string      `json:"ReportingArea"`
	StateCode     string      `json:"StateCode"`
	Latitude      float64     `json:"Latitude"`
	Longitude     float64     `json:"Longitude"`
	ParameterName string      `json:"ParameterName"`
	AQI           int         `json:"AQI"`
	Category      AirCategory `json:"Category"`
	ActionDay     bool        `json:"ActionDay"`
	Discussion    string      `json:"Discussion"`
}
