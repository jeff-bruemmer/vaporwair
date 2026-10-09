package report

import (
	"math"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// JSON output for -json. Each report's data function returns the same objects its text
// shows, under keys shared by every report. Key names carry their units, so a number's
// meaning is never a guess. Scripts depend on these shapes: add keys, never rename them.

// HourJSON is one hour of the hourly forecast.
type HourJSON struct {
	Time        string   `json:"time"`
	TempF       float64  `json:"temp_f"`
	FeelsLikeF  float64  `json:"feels_like_f"`
	DewpointF   *float64 `json:"dewpoint_f"`
	PrecipPct   float64  `json:"precip_pct"`
	PrecipType  string   `json:"precip_type"`
	HumidityPct float64  `json:"humidity_pct"`
	WindMph     float64  `json:"wind_mph"`
	GustMph     *float64 `json:"gust_mph"`
	WindDir     string   `json:"wind_dir"`
	Conditions  string   `json:"conditions"`
}

// DayJSON is one daily period. A night period has no daytime high.
type DayJSON struct {
	Name         string   `json:"name"`
	Start        string   `json:"start"`
	Night        bool     `json:"night"`
	LowF         float64  `json:"low_f"`
	HighF        *float64 `json:"high_f"`
	PrecipPct    float64  `json:"precip_pct"`
	PrecipType   string   `json:"precip_type"`
	WindMph      float64  `json:"wind_mph"`
	GustMph      *float64 `json:"gust_mph"`
	WindDir      string   `json:"wind_dir"`
	Conditions   string   `json:"conditions"`
	Forecast     string   `json:"forecast"`
	PressureInHg *float64 `json:"pressure_inhg"`
	VisibilityMi *float64 `json:"visibility_mi"`
}

// AlertJSON is one active NWS alert.
type AlertJSON struct {
	Event       string  `json:"event"`
	Onset       *string `json:"onset"`
	Expires     *string `json:"expires"`
	Description string  `json:"description"`
}

// AirJSON is one pollutant's forecast for one day. AQI is null until AirNow publishes it.
type AirJSON struct {
	Date      string `json:"date"`
	Pollutant string `json:"pollutant"`
	AQI       *int   `json:"aqi"`
	Category  string `json:"category"`
}

// AQITodayJSON is today's highest AQI.
type AQITodayJSON struct {
	AQI       int    `json:"aqi"`
	Pollutant string `json:"pollutant"`
	Category  string `json:"category"`
}

// OutfitJSON is the clothing recommendation: the outfit is for the coldest it will feel
// before Until.
type OutfitJSON struct {
	Outfit        string   `json:"outfit"`
	FeelsLikeLowF float64  `json:"feels_like_low_f"`
	FeelsLikeAt   *string  `json:"feels_like_low_at"`
	HighF         float64  `json:"high_f"`
	Until         string   `json:"until"`
	Accessories   []string `json:"accessories"`
	Tips          []string `json:"tips"`
}

// InsightsData is the insights report's data for -json.
func InsightsData(w weather.Forecast, a []air.Forecast) map[string]any {
	return map[string]any{
		"now":       nowJSON(w),
		"today":     todayJSON(w),
		"outfit":    outfitJSON(w, a),
		"hours":     hoursJSON(futureHours(w, DefaultMaxHours)),
		"alerts":    alertsJSON(w.Alerts),
		"aqi_today": aqiTodayJSON(a),
	}
}

// SummaryData is the summary report's data for -json.
func SummaryData(w weather.Forecast, a []air.Forecast) map[string]any {
	return map[string]any{
		"now":       nowJSON(w),
		"today":     todayJSON(w),
		"outfit":    outfitJSON(w, a),
		"alerts":    alertsJSON(w.Alerts),
		"aqi_today": aqiTodayJSON(a),
	}
}

// HourlyData is the hourly report's data for -json.
func HourlyData(w weather.Forecast, a []air.Forecast) map[string]any {
	return map[string]any{"hours": hoursJSON(LimitData(upcomingHours(w.Hourly.Data), DefaultHourlyLimit))}
}

// DaysData is the week and daily reports' data for -json.
func DaysData(w weather.Forecast, a []air.Forecast) map[string]any {
	days := []DayJSON{}
	for _, d := range LimitData(w.Daily.Data, 7) {
		days = append(days, dayJSON(d))
	}
	return map[string]any{"days": days}
}

// AlertsData is the alerts report's data for -json.
func AlertsData(w weather.Forecast, a []air.Forecast) map[string]any {
	return map[string]any{"alerts": alertsJSON(w.Alerts)}
}

// AirData is the air quality report's data for -json.
func AirData(w weather.Forecast, a []air.Forecast) map[string]any {
	rows := []AirJSON{}
	for _, f := range a {
		row := AirJSON{Date: f.DateForecast, Pollutant: f.ParameterName, Category: f.Category.Name}
		if f.AQI >= 0 {
			row.AQI = &f.AQI
		}
		rows = append(rows, row)
	}
	return map[string]any{"air": rows}
}

// ClothingData is the clothing report's data for -json: the outfit and the hours it covers.
func ClothingData(w weather.Forecast, a []air.Forecast) map[string]any {
	rec := GetClothingRecommendation(w, a)
	return map[string]any{
		"outfit": outfitFrom(rec),
		"hours":  hoursJSON(windowHours(w, rec.Until)),
	}
}

func nowJSON(w weather.Forecast) HourJSON {
	return hourJSON(GetCurrentHourlyData(w))
}

// todayJSON is the first daily period, or nil when there is none.
func todayJSON(w weather.Forecast) *DayJSON {
	if len(w.Daily.Data) == 0 {
		return nil
	}
	d := dayJSON(w.Daily.Data[0])
	return &d
}

func hoursJSON(hours []weather.DataPoint) []HourJSON {
	out := []HourJSON{}
	for _, h := range hours {
		out = append(out, hourJSON(h))
	}
	return out
}

func hourJSON(h weather.DataPoint) HourJSON {
	return HourJSON{
		Time:        unixRFC3339(h.Time),
		TempF:       Round(h.Temperature),
		FeelsLikeF:  Round(h.ApparentTemperature),
		DewpointF:   reported(h.DewPoint, 0), // 0 means NOAA sent none; see dailyFields
		PrecipPct:   Round(ToPercent(h.PrecipProbability)),
		PrecipType:  h.PrecipType,
		HumidityPct: Round(ToPercent(h.Humidity)),
		WindMph:     Round(h.WindSpeed),
		GustMph:     reported(h.WindGust, 0),
		WindDir:     DegreesToCardinal(h.WindBearing),
		Conditions:  h.Summary,
	}
}

func dayJSON(d weather.DataPoint) DayJSON {
	day := DayJSON{
		Name:         d.PeriodName,
		Start:        unixRFC3339(d.Time),
		Night:        IsNightPeriod(d),
		LowF:         Round(d.TemperatureMin),
		PrecipPct:    Round(ToPercent(d.PrecipProbability)),
		PrecipType:   d.PrecipType,
		WindMph:      Round(d.WindSpeed),
		GustMph:      reported(d.WindGust, 0),
		WindDir:      DegreesToCardinal(d.WindBearing),
		Conditions:   d.Summary,
		Forecast:     d.DetailedForecast,
		PressureInHg: reported(d.Pressure, 2),
		VisibilityMi: reported(d.Visibility, 1),
	}
	if !day.Night {
		high := Round(d.TemperatureMax)
		day.HighF = &high
	}
	return day
}

func alertsJSON(alerts []weather.Alert) []AlertJSON {
	out := []AlertJSON{}
	for _, a := range alerts {
		out = append(out, AlertJSON{
			Event:       a.Title,
			Onset:       optionalTime(a.Time),
			Expires:     optionalTime(a.Expires),
			Description: a.Description,
		})
	}
	return out
}

// aqiTodayJSON is today's highest AQI, or nil when there is no numeric forecast yet.
func aqiTodayJSON(a []air.Forecast) *AQITodayJSON {
	aqi, pollutant, category := GetHighestAQIForToday(a)
	if aqi < 0 {
		return nil
	}
	return &AQITodayJSON{AQI: aqi, Pollutant: pollutant, Category: category}
}

func outfitJSON(w weather.Forecast, a []air.Forecast) OutfitJSON {
	return outfitFrom(GetClothingRecommendation(w, a))
}

func outfitFrom(rec ClothingRecommendation) OutfitJSON {
	out := OutfitJSON{
		Outfit:        rec.Outfit,
		FeelsLikeLowF: Round(rec.Coldest),
		FeelsLikeAt:   optionalTime(rec.ColdestAt),
		HighF:         Round(rec.Warmest),
		Until:         rec.Until.Format(time.RFC3339),
		Accessories:   append([]string{}, rec.Accessories...),
		Tips:          []string{},
	}
	for _, n := range rec.Notes {
		out.Tips = append(out.Tips, n.Text)
	}
	return out
}

func unixRFC3339(t float64) string {
	return time.Unix(int64(t), 0).Format(time.RFC3339)
}

// optionalTime formats a Unix time, or returns nil for 0 (missing).
func optionalTime(t float64) *string {
	if t <= 0 {
		return nil
	}
	s := unixRFC3339(t)
	return &s
}

// reported returns v rounded to places decimals, or nil for 0, which NOAA's data
// uses for "not reported".
func reported(v float64, places int) *float64 {
	if v == 0 {
		return nil
	}
	p := math.Pow10(places)
	r := math.Round(v*p) / p
	return &r
}
