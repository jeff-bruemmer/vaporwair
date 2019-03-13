# Vaporwair
Fast weather and air quality reports in your terminal.  

![alt text](https://github.com/jeff-bruemmer/vaporwair/raw/master/anemometer.png "Anemometer")

## Rationale
Most weather reports do not include air quality, and both air quality and weather services require visiting multiple web pages to get detailed information, which is slow. Vaporwair retrieves both forecasts in the terminal as quickly as possible. It’s written in Go, both for Go’s commandline and OS facilities, as well as its concurrency model.

## Design Constraints
- No external libraries.
- Reports must fit in an unmaximized terminal to avoid scrolling.
- Only one report can be run at a time.

## How Vaporwair Works
Weather obtains user’s coordinates via their IP address, calls the Dark Sky and AirNow APIs to get location-based weather and air quality forecasts, then prints one of several reports, specified by a flag.

### On Vaporwair Speed
1. To prevent needless network calls, Vaporwair determines if the user made a call within the last five minutes. If so, Vaporwair assumes the data is still valid, and executes reports using the last stored call. This shortcut assumes the coordinates have not meaningfully changed in the last minute.
2. If the data has expired, Vaporwair kicks off asynchronous API calls to retrieve new forecasts. It makes optimistic calls to the AirNow and Dark Sky APIs using the previous coordinates, and a call to the IP-API to get the current coordinates. 
3. After Vaporwair acquires the updated coordinates from the IP-API, it compares the coordinates to those used for the optimistic calls in step 2. If the coordinates match, the forecast is valid for the location and Vaporwair executes the report. If not... (Step 4).
4. Vaporwair asynchronously calls the APIs with the updated coordinates and executes the report.

## Sample Reports

### Summary
The default report.
```
$ vaporwair
This week:            Light rain today, with high temperatures bottoming out at 59°F on Sunday.
Currently:            Partly Cloudy.
Current Temperature:  61 °F
Min Temperature:      51 °F at 23:00 HH:MM
Max Temperature:      61 °F at 15:00 HH:MM
Humidity:             74 %
Windspeed:            3 mph
Air Quality Index:    33 PM2.5 Good
Precipitation:        69 %
Precip Type:          rain 
Sunrise:              06:15 HH:MM
Sunset:               17:55 HH:MM
```

### Hourly Weather

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

### Weekly Weather
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
```
vaporwair -a
2019-03-07 
==========
Type      AQI       Category  Description
----      ---       --------  -----------
O3        26        1         Good
PM2.5     33        1         Good
PM10      10        1         Good
NO2       23        1         Good
CO        6         1         Good

2019-03-08 
==========
O3        23        1         Good
PM2.5     21        1         Good
PM10      9         1         Good
NO2       23        1         Good
CO        3         1         Good
```

## Setup
1. Get API keys 
Vaporwair requires two free API keys:

- [Dark Sky](https://darksky.net/dev): for weather reports.
- [AirNow](https://docs.airnowapi.org/): for air quality reports from the Environmental Protection Agency.

2. Download and install the Go programming language.

3. Clone this repository.

4. Change to this repository’s directory, and run `go install`. Make sure your terminal has the Go bin directory in its $PATH.

5. Run `vaporwair` and follow prompts to input API Keys. A configuration directory will automatically be created in your home directory, and the standard report will be run. Specify other reports using the flags listed above. To view a list of available flags, type `vaporwair -help`.

## Roadmap
- Improve entry of API keys with confirmation. Possibly a flag to re-enter API keys.
- Add a flag to specify and configure standard international units.
- Once design finalizes, tests, benchmarking, and documentation upkeep.

## License
M.I.T.

[Powered by Dark Sky](https://darksky.net/poweredby/) and [AirNow](https://airnow.gov/).

