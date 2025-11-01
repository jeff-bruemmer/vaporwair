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

**Features:**

- Temperature-based outfit recommendations (10 temperature ranges)
- Context-aware accessories (rain gear, winter wear, sun protection, air quality masks)
- Safety tips for extreme conditions, UV exposure, air quality, and hydration
- Large temperature swing warnings (bring layers)

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

**Comparative Analysis:**

- Warmest/coldest days of the week
- Windiest day (including gusts)
- Highest precipitation probability
- Largest daily temperature swing
- Automatic notable condition alerts

**Time-Based Insights:**

- Best outdoor window (4-hour blocks optimized for temperature, rain, wind)
- Rain windows (periods with ≥50% precipitation)
- High wind periods (≥20 mph threshold)
- Temperature timing for the next 12 hours

## Setup

1. Obtain a free API key from [AirNow](https://docs.airnowapi.org/) for air quality reports from the Environmental Protection Agency.
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
- Reports must fit in an unmaximized terminal to avoid scrolling.
- Only one report can be run at a time.

## Recent Updates

### Version 2.1 (November 2025 - Enhanced Data Utilization)

- **"Feels Like" Temperature** - Heat index and wind chill calculations now displayed
- **Precipitation Type Detection** - Shows whether it's rain, snow, sleet, or mixed
- **Cloud Cover Display** - Added to hourly reports with percentage coverage
- **Automatic Timezone Detection** - No more hardcoded timezones, uses coordinate-based estimation
- **Dew Point** - Added to insights report for comfort assessment
- **Enhanced Test Coverage** - Comprehensive test suite for new features
- **Full NOAA Data Utilization** - Now using ~95% of available NOAA API data (up from ~60%)

### Version 2.0 (NOAA Migration & New Features)

- **NOAA National Weather Service API** - Migrated from Dark Sky to free NOAA API
- **Zip Code Support** - Get weather for any US location with `-zip` flag
- **Wind Gust Data** - Added to hourly, weekly, and summary reports
- **Temperature Trends** - Shows rising/falling temperature indicators
- **Detailed Forecasts** - Rich narrative descriptions from NOAA
- **Clothing Recommendations** - Smart outfit suggestions based on all weather conditions
- **Insights Report** - Comparative weekly analysis and time-based planning tools
- **Improved Error Handling** - Standardized error patterns throughout
- **Optimistic Caching** - Documented 5-minute cache strategy
- **Presentation Layer** - Abstraction between data and reports for easier maintenance

## Data Sources & Limitations

### What Data Is Available

Vaporwair makes full use of NOAA National Weather Service API data:

**Weather Data (NOAA - No API Key Required):**
- ✅ Temperature (actual)
- ✅ "Feels Like" Temperature (heat index/wind chill calculated from temp, humidity, wind)
- ✅ Precipitation probability
- ✅ Precipitation type (rain, snow, sleet, freezing rain, mix)
- ✅ Cloud cover (estimated from forecast descriptions)
- ✅ Wind speed and direction
- ✅ Wind gusts
- ✅ Humidity
- ✅ Dew point
- ✅ Atmospheric pressure (from observation stations)
- ✅ Visibility (from observation stations)
- ✅ Temperature trends (rising/falling/steady)
- ✅ Weather alerts
- ✅ Detailed forecast narratives

**Air Quality Data (AirNow - API Key Required):**
- ✅ AQI (Air Quality Index) for O3, PM2.5, PM10, NO2, CO
- ✅ Category names (Good, Moderate, Unhealthy, etc.)
- ✅ Multi-pollutant tracking

**Location Services:**
- ✅ US zip code lookup
- ✅ IP-based geolocation
- ✅ Automatic timezone detection (based on coordinates)

### Known Limitations

**Data NOT Available from NOAA API:**
- ❌ **UV Index** - NOAA does not provide UV index data. While the code structure supports it, values are always 0.
- ❌ **Sunrise/Sunset Times** - NOAA forecast API doesn't include precise sunrise/sunset times. Would require astronomical calculations or a separate API.
- ❌ **Moon Phase** - Not provided by NOAA API (legacy field from previous Dark Sky integration)
- ❌ **Ozone Levels** - Not provided by NOAA weather API (air quality ozone is available via AirNow)
- ❌ **Minutely Forecast** - NOAA provides hourly forecasts, not minute-by-minute predictions

**Geographic Limitations:**
- 📍 **US-only coverage** - NOAA API only covers United States territories
- 📍 **Timezone estimation** - Uses longitude-based approximation for US timezones (may be slightly inaccurate near timezone boundaries)

**Temporal Limitations:**
- ⏰ **Hourly forecast** - Up to 7 days ahead (156 hours)
- ⏰ **Daily forecast** - Up to 7 days ahead
- ⏰ **Cache TTL** - 5-minute default (configurable)

### Why Some Fields Are Zero

If you see these values as zero, they are expected limitations:
- **UV Index: 0** - NOAA doesn't provide this (consider integrating EPA UV Index API)
- **Moon Phase: 0** - Not available from NOAA
- **Ozone: 0** - Weather ozone not in NOAA API (air quality O3 available via AirNow)

### Data Accuracy Notes

- **Cloud Cover** - Estimated from NOAA's text descriptions (Sunny=0%, Partly Cloudy=50%, Overcast=100%)
- **"Feels Like" Temperature** - Calculated using standard heat index (temp ≥80°F) and wind chill (temp ≤50°F) formulas
- **Precipitation Type** - Extracted from NOAA forecast text (e.g., "Light Snow" → "snow")
- **Timezone** - Estimated from longitude for US locations (±1 hour accuracy near boundaries)

## Roadmap

- Add flag to re-enter API keys
- Support for international units (metric)
- UV Index integration (EPA or separate API)
- Sunrise/sunset calculations (astronomical formulas or API)
- Additional activity recommendations (outdoor sports, gardening, etc.)
- Historical weather comparisons
- Improved timezone detection (full timezone database)

## License

M.I.T.

Powered by [NOAA National Weather Service API](https://www.weather.gov/documentation/services-web-api) and [AirNow](https://airnow.gov/).
