# V A P O R W A I R

Vaporwair delivers fast weather and air quality reports in your terminal.

> Vaporwair uses the free [National Weather Service API](https://www.weather.gov/documentation/services-web-api) for weather forecasts.

## About Vaporwair

Vaporwair is a command line application that combines weather and air quality forecasts into short, actionable reports:

- **Insights** (default) - Today's forecast, current conditions, what to wear, and the next few hours
- **Summary** - A glance of a few lines: now, high and low, outfit, air quality, and alerts
- **Hourly weather** - Hour-by-hour temperature, feels-like, dew point, precipitation, humidity, wind, and gusts
- **Weekly forecast** - 7-day outlook with lows, highs, precipitation, wind, and conditions
- **Daily detail** - Every NOAA field for each of the next 7 days
- **Weather alerts** - Active NWS watches, warnings, and advisories
- **Air quality report** - EPA AirNow data for multiple pollutants
- **Clothing recommendations** - An outfit chosen for the coldest it will feel over the next 12 hours

Output is plain black-and-white ASCII (section titles are bold in a terminal), so it reads the same in any terminal, pipe, or log file. Set `NO_COLOR=1` to turn off bold as well. For scripts, every report can print its data as JSON with `-json`.

## Rationale

Most weather reports do not include air quality, and both air quality and weather services require visiting multiple web pages to get detailed information, which is slow. Vaporwair retrieves both forecasts in the terminal as quickly as possible. It’s written in Go, both for Go’s commandline and OS facilities, as well as its concurrency model.

## Reports

### Insights (default)

```
$ vaporwair
Burlington 05401 | Tue Oct 6, 19:11 EDT

== WEATHER ALERTS ==============================================================
Alert:    ! FROST ADVISORY | Wed 00:00-05:00 (starts in 4h)
What:     Temperatures as low as 33 will result in frost formation.
Details:  vaporwair alerts

== TONIGHT =====================================================================
Mostly clear, with a low around 37. Southwest wind 3 to 9 mph.

== CURRENT CONDITIONS ==========================================================
Temperature:     51F
Low Tonight:     37F
High Wednesday:  61F
Humidity:        44%
Dewpoint:        30F
Wind:            7 mph from NW
Air Quality:     30 AQI (OZONE) - Good

== WHAT TO WAIR ================================================================
Outfit:   Heavy jacket or coat with layers underneath

== NEXT FEW HOURS ==============================================================
Time   Temp  Feels  Wind  Conditions
       F     F      mph
20:00  48    46     5 NW  Clear
21:00  45    45     3 W   Mostly Clear
22:00  43    43     3 W
23:00  42    42     3 SW
00:00  42    42     3 SW
01:00  40    36     5 S   Partly Cloudy
```

Alerts appear first only when NWS has one active for your location: what it is and when it starts or ends. `vaporwair alerts` shows the full text. After about 6pm the first section is "Tonight", and High / Low become tonight's low and tomorrow's high.

Current Conditions lists what you'd decide on now. The precipitation chance appears only when it is 30% or more, and pressure and visibility are in `vaporwair daily`. Next Few Hours uses the same columns as the hourly report; a Precip or Gust column appears when some hour has one.

### Summary (`vaporwair summary`)

A glance in a few lines, sized for a tmux pane or a login message. A tip line appears when there is one worth acting on.

```
$ vaporwair summary
Burlington 05401 | Tue Oct 6, 19:11 EDT

Now:             51F, Clear, wind 7 mph NW
Low Tonight:     37F
High Wednesday:  61F
Outfit:          Heavy jacket or coat with layers underneath
Air Quality:     30 AQI (OZONE) - Good
Alert:           ! FROST ADVISORY | Wed 00:00-05:00 (starts in 4h)
```

### Hourly weather (`vaporwair hourly`)

Hour-by-hour forecast for the next 12 hours. Units are in the second header row. Conditions are shown when they change, and Precip and Gust columns appear when some hour has a value.

```
$ vaporwair hourly
Burlington 05401 | Tue Oct 6, 19:11 EDT

== HOURLY SUMMARY ==============================================================
Time   Temp  Feels  Dew  Precip  Humid  Wind  Conditions
       F     F      F    %       %      mph
19:00  51    51     30   0       44     7 NW  Clear
20:00  48    46     32   0       54     5 NW
21:00  45    45     34   0       65     3 W   Mostly Clear
22:00  43    43     34   0       70     3 W
23:00  42    42     35   0       76     3 SW
00:00  42    42     35   0       76     3 SW
01:00  40    36     34   0       79     5 S   Partly Cloudy
02:00  39    35     33   1       79     6 S
03:00  39    34     33   1       79     7 S
04:00  39    33     33   1       79     8 S   Mostly Cloudy
05:00  38    32     33   1       82     9 S
06:00  38    30     33   6       82     12 S  Slight Chance Rain Showers
```

### Weekly weather (`vaporwair week`)

A night period such as "Tonight" has no daytime high, so its High is `-`.

```
$ vaporwair week
Burlington 05401 | Tue Oct 6, 19:11 EDT

== WEEK AHEAD ==================================================================
Day           Low  High  Precip  Wind   Conditions
              F    F     %       mph
Tonight       37   -     1       3 SW   Mostly Clear
Wednesday     49   61    22      12 S   Slight Chance Rain Showers
Thursday      43   63    54      9 S    Chance Rain Showers
Friday        41   58    18      6 W    Slight Chance Rain Showers
Saturday      46   59    6       5 W    Mostly Sunny
Sunday        52   69    1       16 SW  Mostly Sunny
Columbus Day  49   66    21      12 SW  Slight Chance Rain Showers
```

### Daily detail (`vaporwair daily`)

All available NOAA fields and the full narrative for each of the next 7 days.

```
$ vaporwair daily
Burlington 05401 | Tue Oct 6, 19:11 EDT

== DAILY FORECAST ==============================================================
Tonight
Low:                   37F
Precipitation Chance:  1%
Wind:                  3 mph from SW
Pressure:              30.02 inHg
Visibility:            10 miles
Forecast:              Mostly clear, with a low around 37. Southwest wind 3
                       to 9 mph.

Wednesday
High:                  61F
Low:                   49F
Rain Chance:           22%
Wind:                  12 mph from S
Forecast:              A slight chance of rain showers after 8am. Mostly
                       cloudy, with a high near 61. South wind 12 to 20 mph,
                       with gusts as high as 31 mph. Chance of precipitation
                       is 20%. New rainfall amounts less than a tenth of an
                       inch possible.
```

### Weather alerts (`vaporwair alerts`)

```
$ vaporwair alerts
Burlington 05401 | Tue Oct 6, 19:11 EDT

== WEATHER ALERTS ==============================================================
Active alerts: 1

! FROST ADVISORY
Effective:       Wed Oct 7, 00:00 EDT
Expires:         Wed Oct 7, 05:00 EDT
Starts In:       4 hours, 49 minutes
What:            Temperatures as low as 33 will result in frost formation.
Where:           The Champlain Valley, which includes Eastern Clinton and
                 Eastern Essex Counties in New York and Western Addison, Grand
                 Isle, Western Chittenden, and Western Franklin Counties in
                 Vermont.
When:            From midnight tonight to 7 AM EDT Wednesday.
Impacts:         Frost could harm sensitive outdoor vegetation. Sensitive
                 outdoor plants may be killed if left uncovered.
```

### Air Quality Report (`vaporwair air`)

EPA AirNow data for each pollutant over the next few days. Categories of "Unhealthy for Sensitive Groups" or worse are marked with `!`.

```
$ vaporwair air
Burlington 05401 | Tue Oct 6, 19:11 EDT

== AIR QUALITY FORECAST ========================================================
Day        Pollutant  AQI  Category
Tue Oct 6  OZONE      30   Good
           PM2.5      25   Good
Wed Oct 7  OZONE      35   Good
           PM2.5      40   Good
```

Air quality forecasts may not be available early in the day; AirNow typically publishes them by late morning. Until then, values show as `pending`.

### Clothing Recommendations (`vaporwair clothing`)

The outfit is chosen for the coldest it will _feel_ (wind chill included) for the rest of the day: until midnight, at most 12 hours ahead and at least 3. `>` marks the recommended tier. Tips are ordered by importance, so health warnings such as unhealthy air come first.

```
$ vaporwair clothing
Burlington 05401 | Tue Oct 6, 19:11 EDT

== WHAT TO WAIR TONIGHT ========================================================

Current:      51F
Until 00:00:  up to 51F, feels as cold as 42F at 23:00

  Feels like  Outfit
  85F+        Light, breathable clothing (shorts, t-shirt, tank top)
  75-84F      Summer wear (shorts or light pants, short sleeves)
  65-74F      Light layers (jeans, long sleeves or light sweater)
  55-64F      Moderate layers (pants, sweater or light jacket)
  45-54F      Warm layers (jacket, long sleeves, jeans)
> 35-44F      Heavy jacket or coat with layers underneath
  25-34F      Winter coat, insulated layers, thermal wear
  15-24F      Heavy winter coat, multiple layers, thermal underwear
  <15F        Extreme cold gear, heavy insulation, thermal base layers

== PRECIPITATION ===============================================================
No precipitation expected until 00:00
```

## Setup

Vaporwair runs on Linux and macOS.

1. (Optional) Obtain a free API key from [AirNow](https://docs.airnowapi.org/) for air quality reports from the Environmental Protection Agency.
   - Weather data is provided by NOAA's National Weather Service API and does not require an API key.

2. Install the [Go programming language](https://go.dev/dl/).

3. Install vaporwair:

   ```
   go install github.com/jeff-bruemmer/vaporwair@latest
   ```

   The binary goes in Go's bin directory, `$(go env GOPATH)/bin`; make sure it is in your `$PATH`.

4. Run `vaporwair` in a terminal, and follow the prompt to input the AirNow API key (the prompt is skipped when stdin isn't a terminal). Vaporwair creates its config file, then shows the Insights report. If you skip the AirNow key, everything except air quality still works; add one later in `~/.config/vaporwair/config.json`.

5. (Optional) Make a zip code your default location: `vaporwair -default=05401`. Without one, vaporwair locates you by IP address.

## Usage

```
Usage: vaporwair [report] [options]

Examples:
  vaporwair                    Today's forecast for your location
  vaporwair hourly -zip=10001  The next 12 hours in New York, this once
  vaporwair -default=05401     Make Burlington, VT your default location
  vaporwair summary -json      The summary as JSON, for scripts

Reports:
  insights   Today's forecast, what to wear, and the next few hours (default)
  summary    A few lines: now, high and low, outfit, air, and alerts
  hourly     Hour by hour for the next 12 hours
  week       One line per day for the next 7 days
  daily      Every NOAA field for each of the next 7 days
  alerts     Active weather alerts and warnings
  air        Air quality forecast by pollutant
  clothing   What to wair, with the full outfit scale

Options:
  -zip CODE      Weather for a US zip code, this once
  -default CODE  Make a zip code your default location (-default=ip clears it)
  -current       Use IP-based location this once
  -refresh       Skip the 5-minute cache and fetch fresh forecasts
  -json          Print the report's data as JSON
  -version       Print the version
```

Options can go before or after the report name, with one dash or two (`-zip` or `--zip`). Run `vaporwair help` (or `-help`, `-h`) to see this list. Help and reports print to stdout; notes, warnings, and errors print to stderr. A mistyped report name gets a suggestion (`did you mean "week"?`).

Exit status is 0 on success, 2 for a usage error (such as an unknown report or a zip code that isn't 5 digits), 130 when interrupted with Ctrl-C, and 1 for any other failure, such as a forecast that can't be fetched.

### Location

Each run picks its location in this order:

1. `-current`: your IP-based location, this once
2. `-zip CODE`: that zip code, this once
3. Your default zip code, set with `-default CODE`
4. Your IP-based location

`-zip` and `-current` never change your default. `-default CODE` saves CODE as your default and shows its forecast; `-default=ip` clears the default, so later runs use your IP-based location. Vaporwair prints a note on stderr whenever the default changes.

```bash
$ vaporwair -default=10001     # New York is now the default
$ vaporwair                    # New York
$ vaporwair -zip=90210         # Beverly Hills, this once
$ vaporwair -current           # Your IP-based location, this once
$ vaporwair                    # New York again
$ vaporwair -default=ip        # Clears the default; IP-based location from now on
```

The header says `(via IP)` when the location came from your IP address rather than a zip code, so a wrong guess is easy to spot.

### JSON output

`-json` prints the selected report's data as JSON on stdout, for scripts and status bars:

```bash
$ vaporwair summary -json | jq -r '"\(.now.temp_f)F \(.now.conditions)"'
57F Chance Rain Showers
```

Every report's JSON has `report`, `generated_at`, `cached_at` (null when just fetched), `offline` (true when fetching failed and the last saved forecast is shown), and `location` (`city`, `zip`, `latitude`, `longitude`, `by_ip`). Each report adds the keys for what it shows:

| Report | Keys |
| --- | --- |
| insights | `now`, `today`, `outfit`, `hours` (next 6), `alerts`, `aqi_today` |
| summary | `now`, `today`, `outfit`, `alerts`, `aqi_today` |
| hourly | `hours` (next 12) |
| week, daily | `days` (next 7) |
| alerts | `alerts` |
| air | `air` |
| clothing | `outfit`, `hours` (the hours the outfit covers) |

The objects are the same in every report. Key names carry their units, times are RFC 3339, and a value that is not reported is `null`.

- **hour** (`now`, `hours`): `time`, `temp_f`, `feels_like_f`, `dewpoint_f`, `precip_pct`, `precip_type`, `humidity_pct`, `wind_mph`, `gust_mph`, `wind_dir`, `conditions`
- **day** (`today`, `days`): `name`, `start`, `night`, `low_f`, `high_f` (null for a night period), `precip_pct`, `precip_type`, `wind_mph`, `gust_mph`, `wind_dir`, `conditions`, `forecast`, `pressure_inhg`, `visibility_mi`
- **alert**: `event`, `onset`, `expires`, `description`
- **air**: `date`, `pollutant`, `aqi` (null until AirNow publishes it), `category`
- **aqi_today**: `aqi`, `pollutant`, `category`, or null when there is no numeric forecast
- **outfit**: `outfit`, `feels_like_low_f`, `feels_like_low_at`, `high_f`, `until`, `accessories`, `tips`

### Files and environment

- `~/.config/vaporwair/config.json`: your AirNow API key and default zip code, readable only by you. `$XDG_CONFIG_HOME` moves it.
- `~/.cache/vaporwair/`: cached forecasts. `$XDG_CACHE_HOME` moves it, and deleting it is always safe.
- `NO_COLOR`: turns off bold section titles.
- `COLUMNS`: the report width when output is piped (at most 80).
- `HTTPS_PROXY`, `NO_PROXY`: route requests through a proxy.

## How Vaporwair works

Vaporwair finds your location (see [Location](#location)), calls the NOAA National Weather Service and AirNow APIs for weather and air quality forecasts, and prints the report you name.

### Services contacted

| Service | When | What it receives |
| --- | --- | --- |
| api.weather.gov (NOAA) | Every fetch | Your location's coordinates |
| www.airnowapi.org (EPA AirNow) | Every fetch, with an API key | Your zip code and API key |
| ipwho.is | Only for IP-based location | Your IP address |
| api.zippopotam.us | When the zip code differs from the last one fetched | The zip code |

Vaporwair sends nothing else, and collects no analytics.

### On Vaporwair speed

1. To prevent needless network calls, Vaporwair uses a 5-minute cache. If you made a call recently for the same location, it serves cached data instead of hitting the APIs again. The header shows how old cached data is. Use `-refresh` to bypass it.

2. When the cache is expired or missing, Vaporwair:
   - Determines your coordinates, reusing the ones saved for your zip code when it has them
   - Asks NOAA for the forecast grid point, then fetches the daily forecast, hourly forecast, alerts, and current observations concurrently, alongside AirNow
   - Gives each request 10 seconds, and the weather forecast 30 seconds in all
   - Displays results once all data is retrieved

3. Cached forecasts are location-aware: asking for a different zip code, or using `-current`, fetches fresh data.

4. If fetching fails (offline, or a service is down), Vaporwair shows the last saved forecast for the same location, up to a day old, and marks the header `(offline)`.

5. While fetching, a small spinner shows on stderr. It only appears when stderr is a terminal, so piped or redirected output stays clean. Ctrl-C stops a fetch at once.

## Design constraints

- Only standard Go packages (i.e. no external libraries).
- Only one report can be run at a time, chosen by name; asking for two is an error.
- Output is black and white and ASCII only.

## Uninstall

```
rm "$(go env GOPATH)/bin/vaporwair"
rm -rf ~/.config/vaporwair ~/.cache/vaporwair
```

## License

M.I.T.

Powered by [NOAA National Weather Service API](https://www.weather.gov/documentation/services-web-api) and [AirNow](https://airnow.gov/).
