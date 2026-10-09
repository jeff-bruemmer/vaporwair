package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// ClothingRecommendation represents a suggested outfit and accessories.
type ClothingRecommendation struct {
	Outfit      string
	Accessories []string
	Notes       []Note // most important first
	// Coldest is the lowest feels-like temperature between now and Until, at ColdestAt
	// (zero when only daily data was available); the outfit is chosen for it, since that
	// is the part of the day you must dress for.
	Coldest   float64
	ColdestAt float64
	Warmest   float64
	Until     time.Time
}

// ClothingHours caps how far ahead clothing recommendations look.
const ClothingHours = 12

// ClothingMinHours is the shortest window, so a late-night check still covers a few hours.
const ClothingMinHours = 3

// Note priorities: lower is more important. ClothingSummary shows only the first note,
// so health and safety advice must sort ahead of comfort tips.
const (
	notePrioritySafety = iota
	notePriorityPrecip
	notePriorityWind
	notePriorityLayers
	notePriorityComfort
)

// Note is one tip. The clothing report gives precipitation tips their own section,
// so it tells them apart by Priority.
type Note struct {
	Priority int
	Text     string
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

// GetClothingRecommendation analyzes weather and air quality (a may be nil) and returns
// clothing suggestions. Temperature, precipitation, and wind all come from the same
// window of hours (now until Until), so the advice is consistent.
func GetClothingRecommendation(f weather.Forecast, a []air.Forecast) ClothingRecommendation {
	var rec ClothingRecommendation
	addNote := func(priority int, format string, args ...any) {
		rec.Notes = append(rec.Notes, Note{priority, fmt.Sprintf(format, args...)})
	}

	rec.Until = clothingWindowEnd(clock())
	hours := windowHours(f, rec.Until)

	// Dress for the coldest it will feel; warm-weather advice uses the warmest it gets.
	// Feels-like already includes wind chill and heat index, so no separate tips for those.
	rec.Coldest, rec.Warmest, rec.ColdestAt = FeelsLikeRange(f, hours)
	temp := rec.Coldest
	hot := rec.Warmest
	rec.Outfit = getBaseOutfit(temp)
	rec.Accessories = []string{}

	// Temperature-based accessories
	if temp < 40 {
		rec.Accessories = append(rec.Accessories, "Winter hat or beanie", "Gloves or mittens", "Scarf")
	}
	if temp < 20 {
		rec.Accessories = append(rec.Accessories, "Face covering or balaclava")
		addNote(notePrioritySafety, "Limit outdoor exposure in extreme cold")
	}

	// Sun protection for hot weather
	if hot >= 75 {
		rec.Accessories = append(rec.Accessories, "Sunglasses", "Sun hat or cap")
		addNote(notePriorityComfort, "Apply sunscreen (SPF 30+)")
	}
	if hot >= 95 {
		addNote(notePrioritySafety, "Avoid strenuous outdoor activity during peak heat")
	}
	if hot >= 85 {
		addNote(notePrioritySafety, "Stay hydrated - bring water bottle")
	}

	// Precipitation: the likeliest hour in the window
	precipProb, precipType := windowPrecip(f, hours)
	if precipProb >= 70 {
		rec.Accessories = append(rec.Accessories, "Umbrella", "Waterproof jacket or raincoat")
		if temp < 45 {
			rec.Accessories = append(rec.Accessories, "Waterproof boots")
		}
		addNote(notePriorityPrecip, "%.0f%% chance of %s", precipProb, precipType)
	} else if precipProb >= 40 {
		addNote(notePriorityPrecip, "%.0f%% chance of %s - consider bringing umbrella", precipProb, precipType)
	}

	// Wind: the strongest sustained speed or gust in the window
	if wind := windowWind(f, hours); wind >= 25 {
		addNote(notePriorityWind, "Very windy, up to %.0f %s - secure loose items", wind, windSpeedUnit)
	} else if wind >= 15 {
		addNote(notePriorityWind, "Breezy, up to %.0f %s", wind, windSpeedUnit)
	}

	// Humidity only changes fabric choice; its effect on temperature is already in feels-like.
	if GetCurrentHourlyData(f).Humidity >= 0.7 && hot >= 75 {
		addNote(notePriorityComfort, "Humid - choose moisture-wicking fabrics")
	}

	// Temperature swing: say when the coldest point is, since that's what the outfit is for
	if rec.Warmest-rec.Coldest >= 15 {
		coldest := fmt.Sprintf("%.0f%s", rec.Coldest, temperatureUnit)
		if rec.ColdestAt > 0 {
			coldest += " at " + FormatTime(rec.ColdestAt)
		}
		addNote(notePriorityLayers, "Dress in layers - feels like %s, up to %.0f%s", coldest, rec.Warmest, temperatureUnit)
	}

	// Air quality: one note per category, advice rather than gear
	if aqi, particle, category := GetHighestAQIForToday(a); aqi >= 0 {
		switch category {
		case "Unhealthy for Sensitive Groups":
			addNote(notePrioritySafety, "Air unhealthy for sensitive groups (%s) - limit prolonged outdoor exertion", particle)
		case "Unhealthy":
			addNote(notePrioritySafety, "Unhealthy air (%s) - limit prolonged outdoor exertion; consider a mask", particle)
		case "Very Unhealthy":
			addNote(notePrioritySafety, "Very unhealthy air (%s) - avoid outdoor activity; wear an N95 outside", particle)
		case "Hazardous":
			addNote(notePrioritySafety, "Hazardous air (%s) - stay indoors; wear an N95 for any outdoor exposure", particle)
		}
	}

	sort.SliceStable(rec.Notes, func(i, j int) bool { return rec.Notes[i].Priority < rec.Notes[j].Priority })
	return rec
}

// windowPrecip returns the highest precipitation chance (as a percent) in hours and its
// type, falling back to the first daily period when there are no hours.
func windowPrecip(f weather.Forecast, hours []weather.DataPoint) (float64, string) {
	if len(hours) == 0 && len(f.Daily.Data) > 0 {
		hours = f.Daily.Data[:1]
	}
	prob, kind := 0.0, ""
	for _, h := range hours {
		if h.PrecipProbability > prob {
			prob, kind = h.PrecipProbability, h.PrecipType
		}
	}
	if kind == "" {
		kind = "precipitation"
	}
	return ToPercent(prob), kind
}

// windowWind returns the strongest sustained wind or gust in hours,
// falling back to the current hour when there are no hours.
func windowWind(f weather.Forecast, hours []weather.DataPoint) float64 {
	if len(hours) == 0 {
		hours = []weather.DataPoint{GetCurrentHourlyData(f)}
	}
	wind := 0.0
	for _, h := range hours {
		wind = max(wind, h.WindSpeed, h.WindGust)
	}
	return wind
}

// ClothingSummary prints a condensed clothing recommendation for the default report.
func ClothingSummary(w weather.Forecast, a []air.Forecast) {
	rec := GetClothingRecommendation(w, a)

	fmt.Println(Title("What to Wair"))

	// Base outfit
	fmt.Fprintf(TW, "Outfit:\t%s\n", rec.Outfit)

	// Accessories (show up to 3 most important)
	if len(rec.Accessories) > 0 {
		count := len(rec.Accessories)
		if count > 3 {
			count = 3
		}
		accessoryList := strings.Join(sentenceList(rec.Accessories[:count]), ", ")
		if len(rec.Accessories) > 3 {
			accessoryList += fmt.Sprintf(", +%d more", len(rec.Accessories)-3)
		}
		fmt.Fprintf(TW, "Bring:\t%s\n", accessoryList)
	}

	// Show most important tip (first one)
	if len(rec.Notes) > 0 {
		fmt.Fprintf(TW, "Tip:\t%s\n", rec.Notes[0].Text)
	}

	// Point to the full report when there's more than fits here
	if len(rec.Notes) > 1 || len(rec.Accessories) > 3 {
		fmt.Fprintf(TW, "More:\tvaporwair clothing\n")
	}
	TW.Flush()
}

// sentenceList lowercases every item after the first, so a joined list reads as one phrase:
// "Winter hat or beanie, gloves or mittens, scarf".
func sentenceList(items []string) []string {
	out := make([]string, len(items))
	for i, item := range items {
		if i > 0 && len(item) > 0 && item[0] >= 'A' && item[0] <= 'Z' {
			item = string(item[0]+32) + item[1:]
		}
		out[i] = item
	}
	return out
}

// periodLabel is the first daily period's name ("This Afternoon", "Tonight"), or "Today".
func periodLabel(w weather.Forecast) string {
	if len(w.Daily.Data) > 0 && w.Daily.Data[0].PeriodName != "" {
		return w.Daily.Data[0].PeriodName
	}
	return "Today"
}

// ClothingReport prints a "What to Wair" recommendation report.
func ClothingReport(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("What to Wair " + periodLabel(w)))
	fmt.Println()

	rec := GetClothingRecommendation(w, a)

	// Find the most current hourly data point (closest to now)
	current := GetCurrentHourlyData(w)

	// Temperature summary
	FormatTemperatureWithFeelsLike(TW, "Current", Round(current.Temperature), Round(current.ApparentTemperature), temperatureUnit)
	// Warmest is an actual temperature and Coldest a feels-like one, so label each.
	coldest := fmt.Sprintf("feels as cold as %.0f%s", Round(rec.Coldest), temperatureUnit)
	if rec.ColdestAt > 0 {
		coldest += " at " + FormatTime(rec.ColdestAt)
	}
	fmt.Fprintf(TW, "Until %s:\tup to %.0f%s, %s\n", rec.Until.Format("15:04"), Round(rec.Warmest), temperatureUnit, coldest)
	TW.Flush()
	fmt.Println()

	// Outfit tiers, with the recommended one marked
	basis := int(rec.Coldest)
	printTableHeader([]column{{"  Feels like", ""}, {"Outfit", ""}})
	for _, level := range OutfitLevels {
		indicator := "  "
		if basis >= level.MinTemp && basis <= level.MaxTemp {
			indicator = "> "
		}

		var tempRange string
		switch {
		case level.MaxTemp >= 999:
			tempRange = fmt.Sprintf("%d%s+", level.MinTemp, temperatureUnit)
		case level.MinTemp <= -100:
			tempRange = fmt.Sprintf("<%d%s", level.MaxTemp+1, temperatureUnit)
		default:
			tempRange = fmt.Sprintf("%d-%d%s", level.MinTemp, level.MaxTemp, temperatureUnit)
		}

		fmt.Fprintf(Table, "%s%s\t%s\n", indicator, tempRange, level.Outfit)
	}
	Table.Flush()
	fmt.Println()

	// Accessories
	if len(rec.Accessories) > 0 {
		fmt.Println("Bring:")
		for _, accessory := range rec.Accessories {
			fmt.Printf("  - %s\n", accessory)
		}
	}

	// Other tips (precipitation tips are covered in their own section below)
	var tips []string
	for _, note := range rec.Notes {
		if note.Priority != notePriorityPrecip {
			tips = append(tips, note.Text)
		}
	}
	if len(tips) > 0 {
		fmt.Println("Tips:")
		for _, tip := range tips {
			for i, line := range WrapText(tip, ReportWidth-4) {
				if i == 0 {
					fmt.Printf("  - %s\n", line)
				} else {
					fmt.Printf("    %s\n", line)
				}
			}
		}
	}

	// Precipitation over the same hours as the outfit, so it explains any rain gear above
	fmt.Println()
	fmt.Println(Title("Precipitation"))
	hours := windowHours(w, rec.Until)
	until := rec.Until.Format("15:04")
	if prob, kind := windowPrecip(w, hours); prob > 0 {
		fmt.Fprintf(TW, "Until %s:\t%.0f%% chance of %s\n", until, prob, kind)
		peak := false
		for _, hour := range hours {
			if hour.PrecipProbability < PrecipSignificantThreshold {
				continue
			}
			if !peak {
				fmt.Fprintf(TW, "Peak times:\n")
				peak = true
			}
			label := hour.PeriodName
			if label == "" {
				label = FormatTime(hour.Time)
			}
			kind := hour.PrecipType
			if kind == "" {
				kind = "precip"
			}
			fmt.Fprintf(TW, "  %-8s  %.0f%% %s\n", label, ToPercent(hour.PrecipProbability), kind)
		}
	} else {
		fmt.Fprintf(TW, "No precipitation expected until %s\n", until)
	}

	TW.Flush()
}
