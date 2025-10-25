package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"strings"
)

// ClothingRecommendation represents a suggested outfit and accessories.
type ClothingRecommendation struct {
	Outfit      string
	Accessories []string
	Notes       []string
}

// getBaseOutfit returns clothing recommendations based on temperature.
func getBaseOutfit(temp float64) string {
	switch {
	case temp >= 85:
		return "Light, breathable clothing (shorts, t-shirt, tank top)"
	case temp >= 75:
		return "Summer wear (shorts or light pants, short sleeves)"
	case temp >= 65:
		return "Light layers (jeans, long sleeves or light sweater)"
	case temp >= 55:
		return "Moderate layers (pants, sweater or light jacket)"
	case temp >= 45:
		return "Warm layers (jacket, long sleeves, jeans)"
	case temp >= 35:
		return "Heavy jacket or coat with layers underneath"
	case temp >= 25:
		return "Winter coat, insulated layers, thermal wear"
	case temp >= 15:
		return "Heavy winter coat, multiple layers, thermal underwear"
	default:
		return "Extreme cold gear, heavy insulation, thermal base layers"
	}
}

// GetClothingRecommendation analyzes weather and returns clothing suggestions.
func GetClothingRecommendation(f weather.Forecast) ClothingRecommendation {
	var rec ClothingRecommendation

	// Get current and daily data
	current := f.Currently
	daily := f.Daily.Data[0]

	// Use the high temp for outfit recommendation
	temp := daily.TemperatureMax
	rec.Outfit = getBaseOutfit(temp)

	rec.Accessories = []string{}
	rec.Notes = []string{}

	// Temperature-based accessories
	if temp < 40 {
		rec.Accessories = append(rec.Accessories, "Winter hat or beanie")
		rec.Accessories = append(rec.Accessories, "Gloves or mittens")
		rec.Accessories = append(rec.Accessories, "Scarf")
	}

	if temp < 20 {
		rec.Accessories = append(rec.Accessories, "Face covering or balaclava")
		rec.Notes = append(rec.Notes, "Limit outdoor exposure in extreme cold")
	}

	// Sun protection for hot weather
	if temp >= 75 {
		rec.Accessories = append(rec.Accessories, "Sunglasses")
		rec.Accessories = append(rec.Accessories, "Sun hat or cap")
		rec.Notes = append(rec.Notes, "Apply sunscreen (SPF 30+)")
	}

	if temp >= 85 {
		rec.Notes = append(rec.Notes, "Stay hydrated - bring water bottle")
	}

	if temp >= 95 {
		rec.Notes = append(rec.Notes, "Avoid strenuous outdoor activity during peak heat")
	}

	// Precipitation recommendations
	precipProb := daily.PrecipProbability * 100
	if precipProb >= 70 {
		rec.Accessories = append(rec.Accessories, "Umbrella")
		rec.Accessories = append(rec.Accessories, "Waterproof jacket or raincoat")
		if temp < 45 {
			rec.Accessories = append(rec.Accessories, "Waterproof boots")
		}
		rec.Notes = append(rec.Notes, fmt.Sprintf("%.0f%% chance of precipitation", precipProb))
	} else if precipProb >= 40 {
		rec.Notes = append(rec.Notes, fmt.Sprintf("%.0f%% chance of precipitation - consider bringing umbrella", precipProb))
	}

	// Wind recommendations
	windSpeed := current.WindSpeed
	windGust := current.WindGust

	if windGust > 0 && windGust > windSpeed {
		windSpeed = windGust // Use gust for recommendations
	}

	if windSpeed >= 25 {
		rec.Notes = append(rec.Notes, "Very windy - secure loose items")
		if temp < 50 {
			rec.Notes = append(rec.Notes, "Wind chill factor - dress warmer than temperature suggests")
		}
	} else if windSpeed >= 15 {
		rec.Notes = append(rec.Notes, "Moderate winds expected")
	}

	// Humidity recommendations
	humidity := current.Humidity * 100
	if humidity >= 70 && temp >= 75 {
		rec.Notes = append(rec.Notes, "High humidity - feels warmer than actual temperature")
		rec.Notes = append(rec.Notes, "Choose moisture-wicking fabrics")
	}

	// UV Index recommendations
	uvIndex := current.UVIndex
	if uvIndex >= 8 {
		rec.Notes = append(rec.Notes, "Very high UV - minimize midday sun exposure")
	} else if uvIndex >= 6 {
		rec.Notes = append(rec.Notes, "High UV - seek shade during midday hours")
	} else if uvIndex >= 3 {
		rec.Notes = append(rec.Notes, "Moderate UV - sun protection recommended")
	}

	// Air quality recommendations
	// This is a simplified check - full implementation would use air quality data

	// Temperature swing recommendations
	tempSwing := daily.TemperatureMax - daily.TemperatureMin
	if tempSwing >= 20 {
		rec.Notes = append(rec.Notes, "Large temperature swing - bring layers you can remove")
	}

	return rec
}

// GetClothingRecommendationWithAir includes air quality in recommendations.
func GetClothingRecommendationWithAir(f weather.Forecast, a []air.Forecast) ClothingRecommendation {
	rec := GetClothingRecommendation(f)

	// Add air quality recommendations
	if len(a) == 0 {
		return rec
	}

	// Find highest AQI for today
	today := a[0].DateForecast
	aqi := -1
	var particle string
	var category string

	for _, measurement := range a {
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
		return rec // No valid air quality data
	}

	// Add air quality recommendations based on category
	switch category {
	case "Unhealthy for Sensitive Groups":
		rec.Notes = append(rec.Notes, fmt.Sprintf("Air Quality: %s (%s)", category, particle))
		rec.Notes = append(rec.Notes, "Sensitive groups should limit prolonged outdoor exertion")
	case "Unhealthy":
		rec.Notes = append(rec.Notes, fmt.Sprintf("Air Quality: %s (%s)", category, particle))
		rec.Notes = append(rec.Notes, "Everyone should limit prolonged outdoor exertion")
		rec.Accessories = append(rec.Accessories, "Consider wearing a mask outdoors")
	case "Very Unhealthy":
		rec.Notes = append(rec.Notes, fmt.Sprintf("Air Quality: %s (%s)", category, particle))
		rec.Notes = append(rec.Notes, "Avoid outdoor activity if possible")
		rec.Accessories = append(rec.Accessories, "Wear N95 or similar mask if going outside")
	case "Hazardous":
		rec.Notes = append(rec.Notes, fmt.Sprintf("Air Quality: %s (%s)", category, particle))
		rec.Notes = append(rec.Notes, "Stay indoors - health alert!")
		rec.Accessories = append(rec.Accessories, "N95 mask required for any outdoor exposure")
	}

	return rec
}

// ClothingSummary prints a condensed clothing recommendation for the default report.
func ClothingSummary(w weather.Forecast, a []air.Forecast) {
	rec := GetClothingRecommendationWithAir(w, a)

	fmt.Println(Title("What to Wear"))

	// Base outfit
	fmt.Fprintf(TW, "Outfit:\t%s\n", rec.Outfit)

	// Accessories (show up to 3 most important)
	if len(rec.Accessories) > 0 {
		count := len(rec.Accessories)
		if count > 3 {
			count = 3
		}
		accessoryList := strings.Join(rec.Accessories[:count], ", ")
		if len(rec.Accessories) > 3 {
			accessoryList += fmt.Sprintf(", +%d more", len(rec.Accessories)-3)
		}
		fmt.Fprintf(TW, "Bring:\t%s\n", accessoryList)
	}

	// Show most important tip (first one)
	if len(rec.Notes) > 0 {
		fmt.Fprintf(TW, "Tip:\t%s\n", rec.Notes[0])
	}

	TW.Flush()

	// Show how to get full details
	if len(rec.Notes) > 1 || len(rec.Accessories) > 3 {
		fmt.Println("(Run with -c flag for detailed clothing recommendations)")
	}
}

// ClothingReport prints a "What to Wear" (wair) recommendation report.
func ClothingReport(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("What to Wear Today"))
	fmt.Println()

	rec := GetClothingRecommendationWithAir(w, a)

	// Temperature summary
	daily := w.Daily.Data[0]
	fmt.Fprintf(TW, "Temperature Range:\t%.0f°F - %.0f°F\n", daily.TemperatureMin, daily.TemperatureMax)
	fmt.Fprintf(TW, "Current:\t%.0f°F\n", w.Currently.Temperature)
	fmt.Println()

	// Base outfit
	fmt.Fprintf(TW, "Recommended Outfit:\n")
	fmt.Fprintf(TW, "\t%s\n", rec.Outfit)
	fmt.Println()

	// Accessories
	if len(rec.Accessories) > 0 {
		fmt.Fprintf(TW, "Accessories:\n")
		for _, accessory := range rec.Accessories {
			fmt.Fprintf(TW, "\t• %s\n", accessory)
		}
		fmt.Println()
	}

	// Additional notes
	if len(rec.Notes) > 0 {
		fmt.Fprintf(TW, "Additional Tips:\n")
		for _, note := range rec.Notes {
			// Wrap long notes to 70 characters
			wrapped := wrapText(note, 70)
			for i, line := range wrapped {
				if i == 0 {
					fmt.Fprintf(TW, "\t• %s\n", line)
				} else {
					fmt.Fprintf(TW, "\t  %s\n", line)
				}
			}
		}
	}

	TW.Flush()
}

// wrapText wraps text to specified width, breaking on spaces.
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
