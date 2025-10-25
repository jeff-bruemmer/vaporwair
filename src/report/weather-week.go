package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"time"
)

func WeatherWeek(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Weekly Summary"))
	fmt.Println(AddPeriod(w.Daily.Summary))
	fmt.Println(Separator)
	data := LimitData(w.Daily.Data, 7)
	formatTitle := "%s\t%s\t%s\t%s\t%s\t%s\t%s\n"
	formatBody := "%v\t%.0f %s\t%.0f %s\t%.0f %s\t%.0f %s\t%.0f %s\t%s\n"
	fmt.Fprintf(TW, formatTitle, "Day", "Min", "Max", "Precip", "Humidity", "Wind", "Gust")
	fmt.Fprintf(TW, formatTitle, "---", "---", "---", "------", "--------", "----", "----")
	for _, day := range data {
		gustStr := "-"
		if day.WindGust > 0 {
			gustStr = fmt.Sprintf("%.0f %s", day.WindGust, windSpeedUnit)
		}
		fmt.Fprintf(TW, formatBody,
			time.Unix(int64(day.Time), 0).Format("Mon"),
			day.TemperatureMin, temperatureUnit,
			day.TemperatureMax, temperatureUnit,
			ToPercent(day.PrecipProbability), percentUnit,
			ToPercent(day.Humidity), percentUnit,
			day.WindSpeed, windSpeedUnit,
			gustStr,
		)
	}
	TW.Flush()
}
