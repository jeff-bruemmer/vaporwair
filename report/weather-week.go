package report

import (
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/weather"
	"github.com/jeff-bruemmer/vaporwair/air"
)

func WeatherWeek(w weather.Forecast, a []air.Forecast) {
	fmt.Println(w)
	fmt.Println(a)
}
