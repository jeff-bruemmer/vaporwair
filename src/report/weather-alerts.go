package report

import (
	"fmt"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// WeatherAlertsReport displays a dedicated report for active weather alerts.
// This shows all available NOAA alert fields including onset time and expiration.
func WeatherAlertsReport(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Weather Alerts"))

	if len(w.Alerts) == 0 {
		fmt.Println("No active weather alerts for this location.")
		fmt.Println()
		fmt.Println("For the week ahead, run: vaporwair week")
		return
	}

	fmt.Printf("Active alerts: %d\n", len(w.Alerts))
	fmt.Println()

	for i, alert := range w.Alerts {
		if i > 0 {
			fmt.Println()
		}

		fmt.Println(AlertHeadline(alert.Title))

		// Show onset time (when alert begins)
		if alert.Time > 0 {
			onsetTime := time.Unix(int64(alert.Time), 0)
			fmt.Fprintf(TW, "Effective:\t%s\n", onsetTime.Format("Mon Jan 2, 15:04 MST"))
		}

		// Show expiration time
		if alert.Expires > 0 {
			expiryTime := time.Unix(int64(alert.Expires), 0)
			fmt.Fprintf(TW, "Expires:\t%s\n", expiryTime.Format("Mon Jan 2, 15:04 MST"))
		}

		// Before onset, count down to the start; once in effect, to the end
		now := clock()
		if onset := unixOrZero(alert.Time); onset.After(now) {
			fmt.Fprintf(TW, "Starts In:\t%s\n", hoursMinutes(onset.Sub(now)))
		} else if alert.Expires > 0 {
			if left := time.Unix(int64(alert.Expires), 0).Sub(now); left > 0 {
				fmt.Fprintf(TW, "Time Remaining:\t%s\n", hoursMinutes(left))
			}
		}

		// Show URI if available
		if alert.URI != "" {
			fmt.Fprintf(TW, "More Info:\t%s\n", alert.URI)
		}

		printAlertSections(alert.Description, "Time Remaining:")
		TW.Flush()
	}
}

// hoursMinutes renders a duration as "9 hours, 48 minutes", or "48 minutes" under an hour.
func hoursMinutes(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	}
	return fmt.Sprintf("%d minutes", minutes)
}
