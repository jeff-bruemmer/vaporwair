# Vaporwair
Fast weather and air quality reports in your terminal.  

![alt text][anemometer]
[anemometer]: https://github.com/jeff-bruemmer/vaporwair/raw/master/anemometer.svg "Anemometer"


## Rationale
I wanted to be able to be able to pop open a terminal and get air quality and weather reports as quickly as possible. I'm a fan of the text interface and command line programs in general, and the weather makes a good case. I chose to write it in go both for its command line and os facilities as well as its concurrency model.

## How Vaporwair Works
Weather gets your coordinates using your IP address, calls the Dark Sky and AirNow APIs to get location based-weather and air quality reports.

## On Vaporwair Speed
1. To prevent needless network calls, Vaporwair determines if you made a call within the last five minutes. If so, it assumes the data is still valid, and Vaporwair executes reports using the last stored call. This shortcut assumes your coordinates have not meaningfully changed in the last minute.
2. Meanwhile, Vaporwair has already kicked off asynchronous API calls in case the stored data has expired vaporwair needs to retrieve a new forecast. This greedy call bets that the most recently used coordinates are your current coordinates. 
3. Also meanwhile, Vaporwair is calling the ip-api to obtain your current coordinates.
4. After aquiring your coordinates, Vaporwair compares the coordinates to those used for the first, greedy call to the APIs in step 2. If the coordinates match, the forecast is valid for your location and Vaporwair executes the report. If not...
5. Vaporwair calls the APIs with your updated coordinates and executes the report.

## Reports


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

