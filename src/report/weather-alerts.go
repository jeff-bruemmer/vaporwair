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
		fmt.Println("Check back later or use -w for weekly forecast.")
		return
	}

	fmt.Printf("Active alerts: %d\n", len(w.Alerts))
	fmt.Println()

	for i, alert := range w.Alerts {
		if i > 0 {
			fmt.Println()
		}

		fmt.Println(Separator)
		fmt.Fprintf(TW, "Alert %d:\t%s\n", i+1, alert.Title)
		fmt.Println(Separator)

		// Show onset time (when alert begins)
		if alert.Time > 0 {
			onsetTime := time.Unix(int64(alert.Time), 0)
			fmt.Fprintf(TW, "Effective:\t%s\n", onsetTime.Format("Mon Jan 2, 15:04 MST"))
		}

		// Show expiration time
		if alert.Expires > 0 {
			expiryTime := time.Unix(int64(alert.Expires), 0)
			fmt.Fprintf(TW, "Expires:\t%s\n", expiryTime.Format("Mon Jan 2, 15:04 MST"))

			// Show time remaining
			timeRemaining := time.Until(expiryTime)
			if timeRemaining > 0 {
				hours := int(timeRemaining.Hours())
				minutes := int(timeRemaining.Minutes()) % 60
				if hours > 0 {
					fmt.Fprintf(TW, "Time Remaining:\t%d hours, %d minutes\n", hours, minutes)
				} else {
					fmt.Fprintf(TW, "Time Remaining:\t%d minutes\n", minutes)
				}
			}
		}

		// Show URI if available
		if alert.URI != "" {
			fmt.Fprintf(TW, "More Info:\t%s\n", alert.URI)
		}

		TW.Flush()

		// Show description with proper wrapping
		if alert.Description != "" {
			fmt.Println()
			maxWidth := calculateValueColumnWidth("Description")
			wrappedDescription := wrapTextForTabwriter(alert.Description, maxWidth)
			fmt.Fprintf(TW, "Description:\t%s\n", wrappedDescription)
			TW.Flush()
		}
	}

	fmt.Println()
	fmt.Println("Stay safe and follow local emergency guidance.")
}
