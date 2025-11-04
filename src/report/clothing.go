package report

import (
	"fmt"
	"strings"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// ClothingRecommendation represents a suggested outfit and accessories.
type ClothingRecommendation struct {
	Outfit      string
	Accessories []string
	Notes       []string
}

// getBaseOutfit returns clothing recommendations based on temperature.
func getBaseOutfit(temp float64) string {
	tempInt := int(temp)
	for _, level := range OutfitLevels {
		if tempInt >= level.MinTemp && tempInt <= level.MaxTemp {
			return level.Outfit
		}
	}
	// Fallback to extreme cold if somehow no match
	return OutfitDescExtremeCold
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
	aqi, particle, category := GetHighestAQIForToday(a)

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

	fmt.Println(Title("What to Wair"))

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

// ClothingReport prints a "What to Wair" recommendation report.
func ClothingReport(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("What to Wair Today"))
	fmt.Println()

	rec := GetClothingRecommendationWithAir(w, a)
	daily := w.Daily.Data[0]

	// Find the most current hourly data point (closest to now)
	current := GetCurrentHourlyData(w)

	// Temperature summary
	fmt.Fprintf(TW, "Current:\t%.0f°F\n", current.Temperature)
	fmt.Fprintf(TW, "Today's Range:\t%.0f°F - %.0f°F\n", daily.TemperatureMin, daily.TemperatureMax)
	fmt.Fprintf(TW, "Temp Range\tOutfit Type\n")
	fmt.Fprintf(TW, "----------\t-----------\n")

	currentTemp := int(current.Temperature)

	for _, level := range OutfitLevels {
		// Check if this is the recommended outfit for current temp or high temp
		indicator := "  "
		if currentTemp >= level.MinTemp && currentTemp <= level.MaxTemp {
			indicator = "→ "
		}

		// Format temp range
		var tempRange string
		if level.MaxTemp >= 999 {
			tempRange = fmt.Sprintf("%d°F+", level.MinTemp)
		} else {
			tempRange = fmt.Sprintf("%d-%d°F", level.MinTemp, level.MaxTemp)
		}

		fmt.Fprintf(TW, "%s%s\t%s\n", indicator, tempRange, level.Outfit)
	}
	TW.Flush()
	fmt.Println()

	// Accessories
	if len(rec.Accessories) > 0 {
		fmt.Fprintf(TW, "Accessories to bring:\n")
		for _, accessory := range rec.Accessories {
			fmt.Fprintf(TW, "\t• %s\n", accessory)
		}
		fmt.Println()
	}

	// Precipitation forecast
	fmt.Println()
	fmt.Println(Title("Precipitation"))

	// Check daily precipitation
	if daily.PrecipProbability > 0 {
		precipType := "precipitation"
		if daily.PrecipType != "" {
			precipType = daily.PrecipType
		}
		fmt.Fprintf(TW, "Today:\t%.0f%% chance of %s\n", ToPercent(daily.PrecipProbability), precipType)

		// Find hours with highest precipitation probability
		type PrecipHour struct {
			time       string
			prob       float64
			precipType string
		}
		var highPrecipHours []PrecipHour

		// Safely slice hourly data to check next 12 hours
		hoursToCheck := SafeSliceHourly(w.Hourly.Data, DefaultHourlyLimit)
		for _, hour := range hoursToCheck {
			if hour.PrecipProbability >= PrecipSignificantThreshold {
				timeStr := hour.PeriodName
				if timeStr == "" {
					timeStr = FormatTime(hour.Time)
				}
				pType := "precip"
				if hour.PrecipType != "" {
					pType = hour.PrecipType
				}
				highPrecipHours = append(highPrecipHours, PrecipHour{
					time:       timeStr,
					prob:       hour.PrecipProbability,
					precipType: pType,
				})
			}
		}

		if len(highPrecipHours) > 0 {
			fmt.Fprintf(TW, "Peak times:\n")
			for _, ph := range highPrecipHours {
				fmt.Fprintf(TW, "  %-8s  %.0f%% %s\n", ph.time, ToPercent(ph.prob), ph.precipType)
			}
		}
	} else {
		fmt.Fprintf(TW, "No precipitation expected today\n")
	}

	// Other important tips (not precipitation related)
	for _, note := range rec.Notes {
		// Skip precipitation-related notes since we handle those above
		if !IsPrecipitationNote(note) {
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
