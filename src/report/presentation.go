package report

import (
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// ForecastViewModel provides a presentation layer between raw API data and reports.
// It pre-processes and organizes data for easier consumption by report functions.
// This abstraction allows data structures to change without impacting report logic.
type ForecastViewModel struct {
	Weather weather.Forecast
	Air     []air.Forecast
}

// NewForecastViewModel creates a view model from raw forecast data.
// This is the entry point for the presentation layer.
func NewForecastViewModel(w weather.Forecast, a []air.Forecast) ForecastViewModel {
	return ForecastViewModel{
		Weather: w,
		Air:     a,
	}
}

// CurrentConditions provides a structured view of current weather conditions.
type CurrentConditions struct {
	Temperature float64
	Humidity    float64
	WindSpeed   float64
	Summary     string
}

// GetCurrentConditions extracts and organizes current weather data.
func (vm ForecastViewModel) GetCurrentConditions() CurrentConditions {
	return CurrentConditions{
		Temperature: vm.Weather.Currently.Temperature,
		Humidity:    vm.Weather.Currently.Humidity,
		WindSpeed:   vm.Weather.Currently.WindSpeed,
		Summary:     vm.Weather.Currently.Summary,
	}
}

// DailyConditions provides a structured view of daily forecast data.
type DailyConditions struct {
	TemperatureMin   float64
	TemperatureMax   float64
	PrecipProbability float64
	PrecipType       string
	Summary          string
}

// GetTodayConditions extracts and organizes today's forecast data.
func (vm ForecastViewModel) GetTodayConditions() DailyConditions {
	if len(vm.Weather.Daily.Data) == 0 {
		return DailyConditions{}
	}

	today := vm.Weather.Daily.Data[0]
	return DailyConditions{
		TemperatureMin:   today.TemperatureMin,
		TemperatureMax:   today.TemperatureMax,
		PrecipProbability: today.PrecipProbability,
		PrecipType:       today.PrecipType,
		Summary:          vm.Weather.Daily.Summary,
	}
}

// AirQualityData provides a structured view of air quality information.
type AirQualityData struct {
	AQI      int
	Particle string
	Category string
	HasData  bool
}

// GetAirQuality extracts and organizes air quality data for today.
func (vm ForecastViewModel) GetAirQuality() AirQualityData {
	if len(vm.Air) == 0 {
		return AirQualityData{HasData: false}
	}

	today := vm.Air[0].DateForecast
	aqi := -1
	var particle string
	var category string

	for _, measurement := range vm.Air {
		if measurement.DateForecast != today {
			break
		}
		if measurement.AQI < 0 {
			continue
		}
		if measurement.AQI > aqi {
			aqi = measurement.AQI
			particle = measurement.ParameterName
			category = measurement.Category.Name
		}
	}

	if aqi < 0 {
		return AirQualityData{HasData: false}
	}

	return AirQualityData{
		AQI:      aqi,
		Particle: particle,
		Category: category,
		HasData:  true,
	}
}
