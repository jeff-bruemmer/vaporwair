package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/air"
	"github.com/jeff-bruemmer/vaporwair/weather"
	"time"
)

func WeatherWeek(w weather.Forecast, a []air.Forecast) {
	fmt.Println(Title("Weekly Summary"))
	fmt.Println(AddPeriod(w.Daily.Summary))
	fmt.Println(Separator)
	data := LimitData(w.Daily.Data, 7)
	formatTitle := "%s\t%s\t%s\t%s\t%s\t%s\t%s\n"
	formatBody := "%v\t%.0f %s\t%.0f %s\t%.0f %s\t%s\t%.0f %s\t%.0f %s\n"
	fmt.Fprintf(TW, formatTitle, "Day", "Min", "Max", "Precip", "Type", "Humidity", "Wind")
	fmt.Fprintf(TW, formatTitle, "---", "---", "---", "------", "----", "--------", "----")
	for _, day := range data {
		fmt.Fprintf(TW, formatBody,
			time.Unix(int64(day.Time), 0).Format("Mon"),
			day.TemperatureMin, tu,
			day.TemperatureMax, tu,
			ToPercent(day.PrecipProbability), pc,
			day.PrecipType,
			ToPercent(day.Humidity), pc,
			day.WindSpeed, wu,
		)
	}
	TW.Flush()
}
