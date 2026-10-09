package report

import (
	"fmt"
	"strings"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// Summary prints a glance of a few lines: now, high and low, what to wear, air, and alerts.
// It is meant for a tmux pane or login message; Insights has the full picture.
func Summary(w weather.Forecast, a []air.Forecast) {
	current := GetCurrentHourlyData(w)
	now := []string{FormatTemperatureString(Round(current.Temperature), Round(current.ApparentTemperature), temperatureUnit)}
	if current.Summary != "" {
		now = append(now, current.Summary)
	}
	now = append(now, "wind "+FormatWindString(current.WindSpeed, current.WindBearing, current.WindGust, windSpeedUnit, false))
	fmt.Fprintf(TW, formatString, "Now", strings.Join(now, ", "))
	printHighLow(TW, w)

	rec := GetClothingRecommendation(w, a)
	fmt.Fprintf(TW, formatString, "Outfit", rec.Outfit)
	if len(rec.Notes) > 0 {
		fmt.Fprintf(TW, formatString, "Tip", rec.Notes[0].Text)
	}

	aqi, ok := airQualityLine(a)
	if !ok {
		aqi = airQualityStatus(a)
	}
	fmt.Fprintf(TW, formatString, "Air Quality", aqi)

	if len(w.Alerts) > 0 {
		headline := alertHeadlineWithWindow(w.Alerts[0], clock())
		if more := len(w.Alerts) - 1; more > 0 {
			headline += fmt.Sprintf(" (+%d more)", more)
		}
		fmt.Fprintf(TW, formatString, "Alert", headline)
	}
	TW.Flush()
}
