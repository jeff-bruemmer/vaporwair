# Vaporwair
Fast weather and air quality reports in your terminal.

> **NOAA National Weather Service API:** Vaporwair now uses the free [National Weather Service API](https://www.weather.gov/documentation/services-web-api) for weather forecasts. No API key required for weather data!

## About Vaporwair
Vaporwair is a command line application that combines weather and air quality forecasts to produce four reports:

- Summary
- Hourly weather
- Weekly forecast
- Air quality report

## Rationale
Most weather reports do not include air quality, and both air quality and weather services require visiting multiple web pages to get detailed information, which is slow. Vaporwair retrieves both forecasts in the terminal as quickly as possible. It’s written in Go, both for Go’s commandline and OS facilities, as well as its concurrency model.

## Reports

### Summary
The default report includes a brief description of the weather, min and max temps, humidity, air quality index, and more.
```
$ vaporwair
Forecasts fetched in 0.50 seconds.
Fri Oct 24 21:22:08 EDT 2025
Lebanon 03766 | 43.6444 , -72.2455
This week:            Mostly cloudy, with a low around 33. Northwest wind around 0 mph.
Currently:            Mostly Cloudy.
Current Temperature:  46 °F
Min Temperature:      33 °F
Max Temperature:      50 °F
Humidity:             83 %
Windspeed:            0 mph
Air Quality Index:    N/A Forecast not yet available
UV Index:             0
Precipitation:        3 %
Precip Type:
```

**Note:** Air quality forecasts from AirNow may show as "not yet available" early in the day, as forecasts are typically published later. When available, it displays the AQI value, pollutant type, and category (e.g., "55 O3 Moderate").

### Hourly weather
The hourly weather report prints a short description of the forecast, as well as the expected temperature, precipitation, precipitation intensity, and wind speed for the next 12 hours.
```
$ vaporwair -h
Partly cloudy until tomorrow afternoon.

Hour      Temp      Feels Like  Precip    Intensity  Wind
----      ----      ----------  ------    ---------  ----
16:00     61 °F     61 °F       0 %       0.00 mmph  6 mph
17:00     59 °F     59 °F       0 %       0.00 mmph  5 mph
18:00     57 °F     57 °F       0 %       0.00 mmph  5 mph
19:00     55 °F     55 °F       8 %       0.21 mmph  6 mph
20:00     54 °F     54 °F       5 %       0.11 mmph  7 mph
21:00     53 °F     53 °F       7 %       0.27 mmph  6 mph
22:00     52 °F     52 °F       10 %      0.45 mmph  5 mph
23:00     51 °F     51 °F       12 %      0.46 mmph  6 mph
00:00     51 °F     51 °F       11 %      0.42 mmph  6 mph
01:00     50 °F     50 °F       10 %      0.36 mmph  7 mph
02:00     50 °F     47 °F       12 %      0.61 mmph  6 mph
03:00     50 °F     48 °F       6 %       0.15 mmph  6 mph
```

### Weekly weather
The weekly weather report prints a short description of the week's forecast, as well as temperature, precipitation, humidity, and wind speed for the coming 7 days.
```
$ vaporwair -w
Light rain today, with high temperatures bottoming out at 60°F on Sunday.

Day       Min       Max       Precip    Type      Humidity  Wind
---       ---       ---       ------    ----      --------  ----
Thu       51 °F     61 °F     69 %      rain      74 %      3 mph
Fri       49 °F     60 °F     31 %      rain      55 %      6 mph
Sat       47 °F     61 °F     8 %       rain      52 %      1 mph
Sun       49 °F     60 °F     35 %      rain      57 %      2 mph
Mon       47 °F     65 °F     13 %      rain      46 %      1 mph
Tue       47 °F     65 °F     28 %      rain      50 %      1 mph
Wed       50 °F     66 °F     4 %       rain      35 %      7 mph
```

### Air Quality Report
The air quality report prints the air quality index for multiple pollutants (typically O3 and PM2.5) for the next several days.

When forecasts are available:
```
$ vaporwair -a
-- AIR QUALITY FORECAST --

2025-10-24
==========
Type      AQI       Category  Description
----      ---       --------  -----------
O3        26        1         Good
PM2.5     33        1         Good

2025-10-25
==========
O3        23        1         Good
PM2.5     21        1         Good
```

When forecasts are not yet available:
```
$ vaporwair -a
-- AIR QUALITY FORECAST --

Air quality forecasts are not yet available.
AirNow typically publishes forecasts later in the day.
```

## Setup
1. Obtain a free API key from [AirNow](https://docs.airnowapi.org/) for air quality reports from the Environmental Protection Agency.
   - Weather data is provided by NOAA's National Weather Service API and does not require an API key.

2. Download and install the [Go programming language](https://golang.org/).

3. Clone this repository.

4. Navigate to this repository's directory, and run `go install`. Make sure your terminal has the [Go bin directory in its $PATH](https://golang.org/doc/gopath_code.html).

5. Run the `vaporwair` binary, and follow the prompts to input the AirNow API key. Vaporwair will create a configuration directory in your home directory, then execute the Summary report

You can specify other reports using the flags listed above in the Reports section. To view a list of available flags, type `vaporwair -help`.

## How Vaporwair works
Vaporwair obtains users coordinates via their IP address, calls the NOAA National Weather Service and AirNow APIs to get location-based weather and air quality forecasts, then prints one of several reports, specified by a flag.

### On Vaporwair speed
1. To prevent needless network calls, Vaporwair determines if the user made a call within the last five minutes. If so, Vaporwair assumes the data is still valid, and executes reports using the last stored call. This shortcut assumes the coordinates have not meaningfully changed in the last five minutes.

2. If the data has expired, Vaporwair kicks off asynchronous API calls to retrieve new forecasts. It makes optimistic calls to the AirNow and NOAA APIs using the previous coordinates, and a call to the IP-API to get the current coordinates.

3. After Vaporwair acquires the updated coordinates from the IP-API, it compares the updated coordinates to the coordinates used for the optimistic calls in step 2. If the coordinates match, the forecast is valid for the location and Vaporwair executes the report. If not: (Step 4).

4. Vaporwair asynchronously calls the APIs with the updated coordinates, waits for the updated forecasts, executes the Summary (or user-flagged) report, and stores the forecast data for subsequent reports.

## Design constraints
- Only standard Go packages (i.e. no external libraries).
- Reports must fit in an unmaximized terminal to avoid scrolling.
- Only one report can be run at a time.

## Roadmap
- Improve entry of API keys with confirmation, fault-tolerance. Possibly a flag to re-enter API keys.
- Add a flag to specify and configure standard international units.
- Once design finalizes, include tests, benchmarks, and additional documentation.

## License
M.I.T.

Powered by [NOAA National Weather Service API](https://www.weather.gov/documentation/services-web-api) and [AirNow](https://airnow.gov/).

