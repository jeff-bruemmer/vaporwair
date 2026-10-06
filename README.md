# Vaporwair

Fast weather and air quality reports in your terminal.

> Vaporwair uses the free [National Weather Service API](https://www.weather.gov/documentation/services-web-api) for weather forecasts. No API key required for weather data!

## About Vaporwair

Vaporwair is a command line application that combines weather and air quality forecasts into short, actionable reports:

- **Insights** (default) - Today's forecast, current conditions, what to wear, and the next few hours
- **Summary** - Current conditions and air quality in one list, plus what to wear
- **Hourly weather** - Hour-by-hour temperature, feels-like, dew point, precipitation, humidity, wind, and gusts
- **Weekly forecast** - 7-day outlook with lows, highs, precipitation, wind, and conditions
- **Daily detail** - Every NOAA field for each of the next 7 days
- **Weather alerts** - Active NWS watches, warnings, and advisories
- **Air quality report** - EPA AirNow data for multiple pollutants
- **Clothing recommendations** - An outfit chosen for the coldest it will feel over the next 12 hours

Output is plain black-and-white ASCII (section titles are bold in a terminal), so it reads the same in any terminal, pipe, or log file. Set `NO_COLOR=1` to turn off bold as well.

## Rationale

Most weather reports do not include air quality, and both air quality and weather services require visiting multiple web pages to get detailed information, which is slow. Vaporwair retrieves both forecasts in the terminal as quickly as possible. It’s written in Go, both for Go’s commandline and OS facilities, as well as its concurrency model.

## Reports

### Insights (default)

```
$ vaporwair
Burlington 05401 | Tue Oct 6, 19:11 EDT

== WEATHER ALERTS ==============================================================
Alert:    ! FROST ADVISORY | until Wed 05:00 (in 9h)
What:     Temperatures as low as 33 will result in frost formation.
Where:    The Champlain Valley, which includes Eastern Clinton and Eastern Essex
          Counties in New York and Western Addison, Grand Isle, Western
          Chittenden, and Western Franklin Counties in Vermont.
When:     From midnight tonight to 7 AM EDT Wednesday.
Impacts:  Frost could harm sensitive outdoor vegetation. Sensitive outdoor
          plants may be killed if left uncovered.

== TONIGHT =====================================================================
Mostly clear, with a low around 37. Southwest wind 3 to 9 mph.

== CURRENT CONDITIONS ==========================================================
Temperature:           51F
High / Low:            51F / 37F
Precipitation Chance:  1%
Humidity:              44%
Dewpoint:              30F
Wind:                  7 mph from NW
Pressure:              30.02 inHg
Visibility:            10 miles
Air Quality:           30 AQI (OZONE) - Good

== WHAT TO WAIR ================================================================
Outfit:   Winter coat, insulated layers, thermal wear
Bring:    Winter hat or beanie, Gloves or mittens, Scarf
Tip:      Dress in layers - feels like 30F at the coldest, up to 51F

== NEXT FEW HOURS ==============================================================
Time   Temp             Conditions    Wind
20:00  48F              Clear         5 mph NW
21:00  45F              Mostly Clear  3 mph W
22:00  43F                            
23:00  42F                            3 mph SW
00:00  42F                            
01:00  40F (feels 36F)                5 mph S
```

Alerts appear first only when NWS has one active for your location. After about 6pm the first section is "Tonight", and High / Low cover the rest of tonight.

### Summary (`-s`)

```
$ vaporwair -s
Burlington 05401 | Tue Oct 6, 19:11 EDT

== WEATHER ALERTS ==============================================================
Alert:    ! FROST ADVISORY | until Wed 05:00 (in 9h)
What:     Temperatures as low as 33 will result in frost formation.
Where:    The Champlain Valley, which includes Eastern Clinton and Eastern Essex
          Counties in New York and Western Addison, Grand Isle, Western
          Chittenden, and Western Franklin Counties in Vermont.
When:     From midnight tonight to 7 AM EDT Wednesday.
Impacts:  Frost could harm sensitive outdoor vegetation. Sensitive outdoor
          plants may be killed if left uncovered.

Tonight:              Mostly clear, with a low around 37. Southwest wind 3
                      to 9 mph.
Currently:            Clear.
Current Temperature:  51F
Min Temperature:      37F
Max Temperature:      51F
Humidity:             44%
Windspeed:            7 mph from NW
Pressure:             30.02 inHg
Visibility:           10 miles
Air Quality Index:    30 OZONE - Good
Precipitation:        1%

== WHAT TO WAIR ================================================================
Outfit:   Winter coat, insulated layers, thermal wear
Bring:    Winter hat or beanie, Gloves or mittens, Scarf
Tip:      Dress in layers - feels like 30F at the coldest, up to 51F
```

### Hourly weather (`-h`)

Hour-by-hour forecast for the next 12 hours. Units are in the second header row.

```
$ vaporwair -h
Burlington 05401 | Tue Oct 6, 19:11 EDT

== HOURLY SUMMARY ==============================================================
Time   Temp  Feels  Dew  Precip  Humid  Wind  Gust
       F     F      F    %       %      mph   mph
19:00  51    51     30   0       44     7     -
20:00  48    46     32   0       54     5     -
21:00  45    45     34   0       65     3     -
22:00  43    43     34   0       70     3     -
23:00  42    42     35   0       76     3     -
00:00  42    42     35   0       76     3     -
01:00  40    36     34   0       79     5     -
02:00  39    35     33   1       79     6     -
03:00  39    34     33   1       79     7     -
04:00  39    33     33   1       79     8     -
05:00  38    32     33   1       82     9     -
06:00  38    30     33   6       82     12    -
```

### Weekly weather (`-w`)

```
$ vaporwair -w
Burlington 05401 | Tue Oct 6, 19:11 EDT

== WEEK AHEAD ==================================================================
Day           Low  High  Precip  Wind  Gust  Conditions
              F    F     %       mph   mph   
Tonight       37   51    1       3     -     Mostly Clear
Wednesday     49   61    22      12    -     Slight Chance Rain Showers
Thursday      43   63    54      9     -     Chance Rain Showers
Friday        41   58    18      6     -     Slight Chance Rain Showers
Saturday      46   59    6       5     -     Mostly Sunny
Sunday        52   69    1       16    -     Mostly Sunny
Columbus Day  49   66    21      12    -     Slight Chance Rain Showers
```

### Daily detail (`-d`)

All available NOAA fields and the full narrative for each of the next 7 days.

```
$ vaporwair -d
Burlington 05401 | Tue Oct 6, 19:11 EDT

== DAILY FORECAST ==============================================================
Mostly clear, with a low around 37. Southwest wind 3 to 9 mph.

Tonight
High:                  51F
Low:                   37F
Precipitation Chance:  1%
Wind:                  3 mph from SW
Pressure:              30.02 inHg
Visibility:            10 miles
Forecast:              Mostly clear, with a low around 37. Southwest wind 3
                       to 9 mph.

Wednesday
High:         61F
Low:          49F
Rain Chance:  22%
Wind:         12 mph from S
Forecast:     A slight chance of rain showers after 8am. Mostly
              cloudy, with a high near 61. South wind 12 to 20 mph,
              with gusts as high as 31 mph. Chance of precipitation
              is 20%. New rainfall amounts less than a tenth of an
```

### Weather alerts (`-alerts`)

```
$ vaporwair -alerts
Burlington 05401 | Tue Oct 6, 19:11 EDT

== WEATHER ALERTS ==============================================================
Active alerts: 1

! FROST ADVISORY
Effective:       Wed Oct 7, 00:00 EDT
Expires:         Wed Oct 7, 05:00 EDT
Time Remaining:  9 hours, 48 minutes
What:            Temperatures as low as 33 will result in frost formation.
Where:           The Champlain Valley, which includes Eastern Clinton and
                 Eastern Essex Counties in New York and Western Addison, Grand
                 Isle, Western Chittenden, and Western Franklin Counties in
                 Vermont.
When:            From midnight tonight to 7 AM EDT Wednesday.
Impacts:         Frost could harm sensitive outdoor vegetation. Sensitive
                 outdoor plants may be killed if left uncovered.

Stay safe and follow local emergency guidance.
```

### Air Quality Report (`-a`)

EPA AirNow data for each pollutant over the next few days. Categories of "Unhealthy for Sensitive Groups" or worse are marked with `!`.

```
$ vaporwair -a
Burlington 05401 | Tue Oct 6, 19:11 EDT

== AIR QUALITY FORECAST ========================================================
  Pollutant      AQI  Category
  ---------      ---  --------
Tue Oct 6
  OZONE           30  Good
  PM2.5           25  Good
Wed Oct 7
  PM2.5           40  Good
  OZONE           35  Good
```

Air quality forecasts may not be available early in the day; AirNow typically publishes them by late morning. Until then, values show as `pending`.

### Clothing Recommendations (`-c`)

The outfit is chosen for the coldest it will *feel* over the next 12 hours (wind chill included), and `>` marks the recommended tier.

```
$ vaporwair -c
Burlington 05401 | Tue Oct 6, 19:11 EDT

== WHAT TO WAIR TODAY ==========================================================

Current:   51F
Next 12h:  feels like 30F to 51F

  Feels like  Outfit
  ----------  ------
  85F+        Light, breathable clothing (shorts, t-shirt, tank top)
  75-84F      Summer wear (shorts or light pants, short sleeves)
  65-74F      Light layers (jeans, long sleeves or light sweater)
  55-64F      Moderate layers (pants, sweater or light jacket)
  45-54F      Warm layers (jacket, long sleeves, jeans)
  35-44F      Heavy jacket or coat with layers underneath
> 25-34F      Winter coat, insulated layers, thermal wear
  15-24F      Heavy winter coat, multiple layers, thermal underwear
  <15F        Extreme cold gear, heavy insulation, thermal base layers

Bring:
  - Winter hat or beanie
  - Gloves or mittens
  - Scarf
Tips:
  - Dress in layers - feels like 30F at the coldest, up to 51F

== PRECIPITATION ===============================================================
Today:    1% chance of precipitation
```

## Setup

1. (Optional) Obtain a free API key from [AirNow](https://docs.airnowapi.org/) for air quality reports from the Environmental Protection Agency.
   - Weather data is provided by NOAA's National Weather Service API and does not require an API key.

2. Download and install the [Go programming language](https://golang.org/).

3. Clone this repository.

4. Navigate to this repository's directory, and run `go install`. Make sure your terminal has the [Go bin directory in its $PATH](https://golang.org/doc/gopath_code.html).

5. Run the `vaporwair` binary, and follow the prompts to input the AirNow API key. Vaporwair will create a configuration directory in your home directory, then show the Insights report. If you skip the AirNow key, everything except air quality still works; add one later in `~/.vaporwair/config.json`.

## Available Flags

```
  -h            Hourly weather forecast (next 12 hours)
  -w            Weekly weather forecast (next 7 days)
  -d            Daily detail (all NOAA fields for the next 7 days)
  -alerts       Active weather alerts
  -a            Air quality report
  -c            Clothing recommendations (what to wair)
  -i            Insights report (the default)
  -s            Summary report
  -zip CODE     Get weather for a specific US zip code (e.g., -zip=10001)
  -current      Use IP-based location (temporary override, doesn't change default)
  -refresh      Skip the 5-minute cache and fetch fresh forecasts
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

1. To prevent needless network calls, Vaporwair uses a 5-minute cache. If you made a call recently for the same location (including your saved default zip code), it serves cached data instead of hitting the APIs again. The header shows how old cached data is. Use `-refresh` to bypass it.

2. When cache is expired or missing, Vaporwair:
   - Determines your coordinates (via zip code or IP geolocation)
   - Makes parallel async calls to NOAA and AirNow APIs
   - Both API calls execute concurrently with a 30-second timeout
   - Displays results once all data is retrieved

3. Cached forecasts are location-aware - asking for a different zip code, or using `-current`, fetches fresh data.

4. While fetching, a small spinner shows on stderr. It only appears when stderr is a terminal, so piped or redirected output stays clean.

## Design constraints

- Only standard Go packages (i.e. no external libraries).
- Only one report can be run at a time.
- Output is black and white and ASCII only.

## License

M.I.T.

Powered by [NOAA National Weather Service API](https://www.weather.gov/documentation/services-web-api) and [AirNow](https://airnow.gov/).
