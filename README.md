# Vaporwair

Fast weather and air quality reports in your terminal.

> Vaporwair uses the free [National Weather Service API](https://www.weather.gov/documentation/services-web-api) for weather forecasts. No API key required for weather data!

## About Vaporwair

Vaporwair is a command line application that combines weather and air quality forecasts to produce intelligent, actionable reports:

- **Summary** - Overview with "feels like" temp, weather, air quality, and clothing recommendations
- **Hourly weather** - Hour-by-hour forecast with temp, precipitation type, clouds, wind, and gusts
- **Weekly forecast** - 7-day outlook with daily conditions
- **Air quality report** - EPA AirNow data for multiple pollutants
- **Clothing recommendations** - Smart outfit suggestions based on all weather conditions
- **Insights** - Comparative analysis with dew point, pressure, and time-based planning tools

## Rationale

Most weather reports do not include air quality, and both air quality and weather services require visiting multiple web pages to get detailed information, which is slow. Vaporwair retrieves both forecasts in the terminal as quickly as possible. It’s written in Go, both for Go’s commandline and OS facilities, as well as its concurrency model.

## Reports

### Summary (default)

The default report includes weather overview, current conditions, air quality, and clothing recommendations.

```
$ vaporwair
Forecasts fetched in 0.50 seconds.
Fri Oct 24 21:22:08 EDT 2025
Lebanon 03766 | 43.6444 , -72.2455
This week:            Mostly cloudy, with a low around 33. Northwest wind around 0 mph.
Currently:            Mostly Cloudy.
Details:              Mostly cloudy, with a low around 33.
Current Temperature:  46 °F
Min Temperature:      33 °F
Max Temperature:      50 °F
Temp Trend:           falling
Humidity:             83 %
Windspeed:            0 mph
Wind Gust:            5 mph
Air Quality Index:    N/A Forecast not yet available
UV Index:             0
Precipitation:        3 %

-- WHAT TO WEAR --
Outfit:     Moderate layers (pants, sweater or light jacket)
Bring:      Winter hat or beanie, Gloves or mittens
Tip:        Wind chill factor - dress warmer than temperature suggests
(Run with -c flag for detailed clothing recommendations)
```

**New in Summary:**

- **Detailed forecast** - Rich narrative description from NOAA
- **Temperature trend** - Rising, falling, or steady temperatures
- **Wind gusts** - Peak wind speeds beyond sustained winds
- **Clothing recommendations** - Smart outfit suggestions based on all conditions

### Hourly weather (`-h`)

Hour-by-hour forecast for the next 12 hours, including temperature, precipitation, **cloud cover**, wind, and gusts.

```
$ vaporwair -h
-- HOURLY SUMMARY --
Partly cloudy until tomorrow afternoon.

Hour   Temp    Precip  Clouds  Wind    Gust
----   ----    ------  ------  ----    ----
16:00  61 °F   0 %     50 %    6 mph   -
17:00  59 °F   0 %     65 %    5 mph   -
18:00  57 °F   0 %     75 %    5 mph   8 mph
19:00  55 °F   8 %     75 %    6 mph   -
20:00  54 °F   5 %     80 %    7 mph   -
21:00  53 °F   7 %     85 %    6 mph   -
22:00  52 °F   10 %    90 %    5 mph   -
23:00  51 °F   12 %    95 %    6 mph   -
00:00  51 °F   11 %    100 %   6 mph   -
01:00  50 °F   10 %    100 %   7 mph   -
02:00  50 °F   12 %    95 %    6 mph   -
03:00  50 °F   6 %     85 %    6 mph   -
```

### Weekly weather (`-w`)

7-day forecast with daily temperature ranges, precipitation, humidity, wind, and gusts.

```
$ vaporwair -w
-- WEEKLY SUMMARY --
Light rain today, with high temperatures bottoming out at 60°F on Sunday.
+++
Day   Min     Max     Precip  Humidity  Wind    Gust
---   ---     ---     ------  --------  ----    ----
Thu   51 °F   61 °F   69 %    74 %      3 mph   -
Fri   49 °F   60 °F   31 %    55 %      6 mph   -
Sat   47 °F   61 °F   8 %     52 %      1 mph   -
Sun   49 °F   60 °F   35 %    57 %      2 mph   5 mph
Mon   47 °F   65 °F   13 %    46 %      1 mph   -
Tue   47 °F   65 °F   28 %    50 %      1 mph   -
Wed   50 °F   66 °F   4 %     35 %      7 mph   12 mph
```

### Air Quality Report (`-a`)

EPA AirNow data showing air quality index for multiple pollutants over the next several days.

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

Air quality forecasts may not be available early in the day. Forecasts are typically published by late morning.

### Clothing Recommendations (`-c`)

Smart outfit and accessory suggestions based on comprehensive weather analysis.

```
$ vaporwair -c
-- WHAT TO WEAR TODAY --

Temperature Range:  42°F - 50°F
Current:            46°F

Recommended Outfit:
  Warm layers (jacket, long sleeves, jeans)

Accessories:
  • Winter hat or beanie
  • Gloves or mittens
  • Scarf

Additional Tips:
  • Wind chill factor - dress warmer than temperature suggests
  • Moderate winds expected
  • Moderate UV - sun protection recommended
```

### Insights Report (`-i`)

Comparative analysis and time-based planning tools for the week ahead.

```
$ vaporwair -i
-- WEEKLY COMPARISON --

Warmest:              65°F on Tuesday
Coldest:              33°F on Friday
Windiest:             12 mph on Wednesday
Rainiest:             69% chance on Thursday
Biggest Temp Swing:   18°F on Friday

Notable:
  • Wide temperature range this week (32°F difference)
  • High chance of rain on Thursday
  • Large temperature swing on Friday - dress in layers


-- TIME-BASED INSIGHTS --

Best 4-hour outdoor window:
  Sat 1:00 PM - 5:00 PM
  63°F average, no rain expected

Rain windows (50%+ chance):
  Thu 6:00 AM - 11:00 AM

High wind periods (20+ mph):
  Wed 2:00 PM - 6:00 PM

Temperature timing (next 12 hours):
  Warmest: 50°F at 3:00 PM
  Coldest: 42°F at 6:00 AM
```

## Setup

1. (Optional) Obtain a free API key from [AirNow](https://docs.airnowapi.org/) for air quality reports from the Environmental Protection Agency.
   - Weather data is provided by NOAA's National Weather Service API and does not require an API key.

2. Download and install the [Go programming language](https://golang.org/).

3. Clone this repository.

4. Navigate to this repository's directory, and run `go install`. Make sure your terminal has the [Go bin directory in its $PATH](https://golang.org/doc/gopath_code.html).

5. Run the `vaporwair` binary, and follow the prompts to input the AirNow API key. Vaporwair will create a configuration directory in your home directory, then execute the Summary report.

## Available Flags

```
  -h            Hourly weather forecast (next 12 hours)
  -w            Weekly weather forecast (next 7 days)
  -a            Air quality report
  -c            Clothing recommendations (what to wair)
  -i            Insights report (comparative analysis & time-based planning)
  -zip CODE     Get weather for a specific US zip code (e.g., -zip=10001)
  -current      Use IP-based location (temporary override, doesn't change default)
```

Run `vaporwair -help` to see all available options.

### Zip Code Usage

By default, Vaporwair uses your IP address to determine your location. You can get weather for any US location using the `-zip` flag:

```bash
$ vaporwair -zip=10001          # New York, NY
$ vaporwair -zip=90210          # Beverly Hills, CA
$ vaporwair -zip=60601 -h       # Chicago, IL (hourly forecast)
$ vaporwair -zip=33101 -c       # Miami, FL (clothing recommendations)
```

**Default Zip Code Behavior:**
- When you use `-zip`, that zip code is automatically saved as your default location
- Subsequent runs will use the saved zip code instead of IP-based geolocation
- To return to IP-based location permanently, edit `~/.vaporwair/config.json` and remove the `defaultzipcode` field
- Using a different `-zip` flag updates your default to the new location

**Temporary Location Override:**
- Use `-current` to temporarily get weather for your current IP-based location
- This does NOT clear or change your saved default zip code
- Useful for travelers who want to check local conditions without changing their home location

**Example workflow:**
```bash
$ vaporwair -zip=10001     # Sets default to NYC, shows NYC weather
$ vaporwair                # Now shows NYC weather (using saved default)
$ vaporwair -current       # Shows weather for current IP location (default still NYC)
$ vaporwair                # Back to NYC weather (saved default unchanged)
$ vaporwair -zip=90210     # Sets default to LA, shows LA weather
$ vaporwair                # Now shows LA weather (using new default)
```

## How Vaporwair works

Vaporwair obtains coordinates using a priority system:
1. `-current` flag: Uses your current IP-based location (temporary override)
2. `-zip` flag: Uses the specified zip code and saves it as default
3. Saved default zip code: Uses the last zip code you specified
4. IP geolocation: Falls back to IP-based location if no zip code is set

It then calls the NOAA National Weather Service and AirNow APIs to get location-based weather and air quality forecasts, and prints one of several reports specified by flags.

### On Vaporwair speed

1. To prevent needless network calls, Vaporwair determines if the user made a call within the last five minutes. If so, Vaporwair assumes the data is still valid, and executes reports using the last stored call. This shortcut assumes the coordinates have not meaningfully changed in the last five minutes.

2. If the data has expired, Vaporwair kicks off asynchronous API calls to retrieve new forecasts. It makes optimistic calls to the AirNow and NOAA APIs using the previous coordinates, and a call to the IP-API to get the current coordinates.

3. After Vaporwair acquires the updated coordinates from the IP-API, it compares the updated coordinates to the coordinates used for the optimistic calls in step 2. If the coordinates match, the forecast is valid for the location and Vaporwair executes the report. If not: (Step 4).

4. Vaporwair asynchronously calls the APIs with the updated coordinates, waits for the updated forecasts, executes the Summary (or user-flagged) report, and stores the forecast data for subsequent reports.

## Design constraints

- Only standard Go packages (i.e. no external libraries).
- Only one report can be run at a time.

## License

M.I.T.

Powered by [NOAA National Weather Service API](https://www.weather.gov/documentation/services-web-api) and [AirNow](https://airnow.gov/).
