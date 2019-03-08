# Vaporwair
Fast weather and air quality reports in your terminal.  

![alt text](https://github.com/jeff-bruemmer/vaporwair/raw/master/anemometer.png "Anemometer")

## Rationale
I wanted to be able to be able to pop open a terminal and get air quality and weather reports as quickly as possible. I'm a fan of the text interface and command line programs in general, and the weather makes a good case. I chose to write it in go both for its command line and os facilities as well as its concurrency model.

## How Vaporwair Works
Weather gets your coordinates using your IP address, calls the Dark Sky and AirNow APIs to get location based-weather and air quality forecasts, then prints one of several reports, specified by a flag.

## On Vaporwair Speed
1. To prevent needless network calls, Vaporwair determines if you made a call within the last five minutes. If so, it assumes the data is still valid, and Vaporwair executes reports using the last stored call. This shortcut assumes your coordinates have not meaningfully changed in the last minute.
2. Meanwhile, Vaporwair has already kicked off asynchronous API calls in case the stored data has expired vaporwair needs to retrieve a new forecast. This greedy call bets that the most recently used coordinates are your current coordinates. 
3. Also meanwhile, Vaporwair is calling the ip-api to obtain your current coordinates.
4. After aquiring your coordinates, Vaporwair compares the coordinates to those used for the first, greedy call to the APIs in step 2. If the coordinates match, the forecast is valid for your location and Vaporwair executes the report. If not...
5. Vaporwair calls the APIs with your updated coordinates and executes the report.

## Design Constraints
- No external libraries.
- Reports must fit in unmaximized terminal to avoid the need to scroll up to read beginning of report.
- Only one report can be run at a time.

## Reports

### Summary
The default report.
```
Summary:              Partly Cloudy
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
Vaporwair requires two free API keys.

- [Dark Sky](https://darksky.net/dev): for weather reports.
- [AirNow](https://docs.airnowapi.org/): for air quality reports from the Environmental Protection Agency.

2. Create a config file called ".vaporwair-config.json"

3. Save keys to your config file. The file should contain the following json string:

```
{"darkskyapikey": "YOUR_DARK_SKY_API_KEY_HERE",
 "airnowapikey": "YOUR_AIRNOW_API_KEY_HERE"}
```
Substitute your api keys for the values. 

## Reports
Run default weather and air quality report: run with no arguments.
```
> vaporwair
```

[Powered by Dark Sky](https://darksky.net/poweredby/) and [AirNow](https://airnow.gov/).

