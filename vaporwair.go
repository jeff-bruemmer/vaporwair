package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"github.com/jeff-bruemmer/vaporwair/src/report"
	"github.com/jeff-bruemmer/vaporwair/src/storage"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"time"
)

// reportSpec is one report: its subcommand and a one-line description for usage.
type reportSpec struct {
	name, desc string
	run        func(weather.Forecast, []air.Forecast)
}

// reports lists every report; the first is the default.
var reports = []reportSpec{
	{"insights", "Today's forecast, what to wear, and the next few hours", report.InsightsReport},
	{"summary", "A few lines: now, high and low, outfit, air, and alerts", report.Summary},
	{"hourly", "Hour by hour for the next 12 hours", report.WeatherHourly},
	{"week", "One line per day for the next 7 days", report.WeatherWeek},
	{"daily", "Every NOAA field for each of the next 7 days", report.WeatherDaily},
	{"alerts", "Active weather alerts and warnings", report.WeatherAlertsReport},
	{"air", "Air quality forecast by pollutant", report.AirQuality},
	{"clothing", "What to wair, with the full outfit scale", report.ClothingReport},
}

// selected is the report this run prints.
var selected = reports[0]

// Location and cache options
var zipCode string
var useCurrentLocation bool
var refresh bool

// forgetZip is set by -zip=ip: clear the saved default and use IP-based location.
var forgetZip bool

// ipZip is the -zip value that clears the saved default.
const ipZip = "ip"

// options are the flags shown in usage, in order.
var options = []string{"zip", "current", "refresh"}

// notes collects non-fatal warnings raised while fetching.
// They're printed after the report so they don't garble the spinner line.
// Only the main goroutine appends to it.
var notes []string

// stopSpinner clears the spinner if one is running. fatalf calls it so an error
// message never lands on the same line as the spinner.
var stopSpinner = func() {}

// fatalf stops the spinner, prints an error to stderr, and exits.
func fatalf(format string, args ...any) {
	stopSpinner()
	log.Fatalf(format, args...)
}

// isValid checks if cached forecast is still fresh (optimistic caching).
func isValid(t time.Time, timeout float64) bool {
	return time.Since(t).Minutes() < timeout
}

// startSpinner shows an ASCII spinner and label on stderr until stopSpinner is called.
// It draws nothing when stderr isn't a terminal, so piped or redirected output stays clean.
func startSpinner(label string) {
	if !report.IsTerminal(os.Stderr) {
		return
	}
	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		frames := `|/-\`
		tick := time.NewTicker(100 * time.Millisecond)
		defer tick.Stop()
		for i := 0; ; i++ {
			fmt.Fprintf(os.Stderr, "\r%c %s", frames[i%len(frames)], label)
			select {
			case <-done:
				fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", len(label)+2))
				return
			case <-tick.C:
			}
		}
	}()
	stopSpinner = func() {
		close(done)
		<-finished
		stopSpinner = func() {}
	}
}

// locatedByIP is true when this run's location came from IP lookup rather than a zip code.
// The header says so, since an IP can place you far from where you are.
var locatedByIP bool

// getIPGeoData retrieves geolocation data from IP address.
func getIPGeoData() (geolocation.Coordinates, error) {
	geoData, err := geolocation.GetGeoData(geolocation.IPAPIAddress)
	if err != nil {
		return geolocation.Coordinates{}, fmt.Errorf("Failed to determine location from IP address: %w\nTry specifying a location with -zip <code>.", err)
	}
	locatedByIP = true
	return geolocation.FormatCoordinates(geoData), nil
}

// GetCoordinates retrieves coordinates based on zip code flag, default zip, or IP address.
// Priority: 1) -current flag (IP), 2) -zip flag, 3) default zip from config, 4) IP geolocation
func GetCoordinates(appConfig storage.AppConfig) (geolocation.Coordinates, error) {
	// Force IP-based location (temporarily override default zip)
	if useCurrentLocation {
		return getIPGeoData()
	}

	// Get coordinates from zip code flag
	if zipCode != "" {
		geoData, err := geolocation.GetGeoDataFromZip(zipCode)
		if err != nil {
			return geolocation.Coordinates{}, fmt.Errorf("Failed to get location for zip code %s: %w\nPlease verify the zip code is valid.", zipCode, err)
		}
		return geolocation.FormatCoordinates(geoData), nil
	}

	// Try saved default zip code with fallback to IP
	if appConfig.Config.DefaultZipCode != "" {
		geoData, err := geolocation.GetGeoDataFromZip(appConfig.Config.DefaultZipCode)
		if err != nil {
			coords, ipErr := getIPGeoData()
			if ipErr != nil {
				return coords, fmt.Errorf("Failed to look up default zip code %s: %w", appConfig.Config.DefaultZipCode, err)
			}
			notes = append(notes, fmt.Sprintf("Could not use default zip code %s (%v); used IP-based location", appConfig.Config.DefaultZipCode, err))
			return coords, nil
		}
		return geolocation.FormatCoordinates(geoData), nil
	}

	// Default: use IP geolocation
	return getIPGeoData()
}

func PrintBanner() {
	banner := "\nV A P O R W A I R\n"
	fmt.Fprintln(os.Stderr, banner)
}

// PrintHeader prints a one-line header: location, time, and cache age when served from cache.
// A zero cachedAt means the forecast was just fetched; stale marks an expired cache shown
// because fetching failed. byIP marks a location found from the IP address.
func PrintHeader(t time.Time, c geolocation.Coordinates, cachedAt time.Time, stale, byIP bool) {
	parts := []string{}
	if place := strings.TrimSpace(c.City + " " + c.Zip); place != "" {
		if byIP {
			place += " (via IP)"
		}
		parts = append(parts, report.Bold(place))
	}
	parts = append(parts, t.Format("Mon Jan 2, 15:04 MST"))
	if !cachedAt.IsZero() {
		age := t.Sub(cachedAt)
		switch {
		case age < time.Minute:
			parts = append(parts, "cached just now")
		case age < time.Hour:
			parts = append(parts, fmt.Sprintf("cached %dm ago", int(age.Minutes())))
		default:
			parts = append(parts, fmt.Sprintf("cached %dh ago", int(age.Hours())))
		}
		if stale {
			parts[len(parts)-1] += " (offline)"
		}
	}
	fmt.Println(strings.Join(parts, " | "))
	fmt.Println()
}

// PrintNotes prints any warnings collected while fetching, after the report.
func PrintNotes() {
	if len(notes) > 0 {
		fmt.Println()
	}
	for _, n := range notes {
		fmt.Println("Note: " + n)
	}
}

// SaveForecasts persists forecasts to disk with timestamp and coordinates.
func SaveForecasts(homeDir string, coordinates geolocation.Coordinates, wf weather.Forecast, af []air.Forecast) {
	storage.UpdateLastCall(coordinates, homeDir+storage.SavedCallFileName)
	storage.SaveWeatherForecast(homeDir+storage.SavedWeatherFileName, wf)
	storage.SaveAirForecast(homeDir+storage.SavedAirFileName, af)
}

// newFlagSet defines the options. The flag package's own error and usage output is off:
// main prints a short error, or the full usage for -help.
func newFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("vaporwair", flag.ContinueOnError)
	fs.StringVar(&zipCode, "zip", "", "Weather for a US zip code, saved as the default (-zip=ip clears it)")
	fs.BoolVar(&useCurrentLocation, "current", false, "Use IP-based location this once (default unchanged)")
	fs.BoolVar(&refresh, "refresh", false, "Skip the 5-minute cache and fetch fresh forecasts")
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	return fs
}

// printUsage lists the reports and options.
func printUsage(fs *flag.FlagSet) {
	PrintBanner()
	fmt.Fprintf(os.Stderr, "Usage: vaporwair [report] [options]\n\nReports:\n")
	for i, r := range reports {
		desc := r.desc
		if i == 0 {
			desc += " (default)"
		}
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", r.name, desc)
	}
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	for _, name := range options {
		f := fs.Lookup(name)
		arg := ""
		if name == "zip" {
			arg = " CODE"
		}
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", "-"+name+arg, f.Usage)
	}
	fmt.Fprintf(os.Stderr, "\nExample: vaporwair hourly -zip=10001\n")
}

// setupConfiguration initializes the configuration directory and loads API keys.
// Returns the application configuration or an error if setup fails.
func setupConfiguration() (storage.AppConfig, error) {
	appConfig, err := storage.InitializeAppConfig()
	if err != nil {
		return appConfig, err
	}

	// Explain a missing AirNow key once, at setup. Later runs get a one-line note
	// only on reports that show air quality.
	if appConfig.Config.AirNowAPIKey == "" {
		configFile := appConfig.HomeDir + storage.ConfigFileName
		if appConfig.FirstRun {
			fmt.Fprintln(os.Stderr, "No AirNow API key: air quality data will not be available.")
			fmt.Fprintln(os.Stderr, "Get a free key at https://docs.airnowapi.org/account/request/")
			fmt.Fprintln(os.Stderr, "and add it to "+configFile)
			fmt.Fprintln(os.Stderr)
		} else if selected.name == "air" || selected.name == "summary" {
			notes = append(notes, "Air quality off: no AirNow API key in "+configFile)
		}
	}

	return appConfig, nil
}

// fetchForecasts retrieves weather and air quality forecasts for given coordinates.
// Returns weather forecast and air quality forecast.
func fetchForecasts(coords geolocation.Coordinates, config storage.Config) (weather.Forecast, []air.Forecast, error) {
	type airResult struct {
		forecast []air.Forecast
		note     string // non-empty when AirNow failed; reported, not fatal
	}
	weatherChan := make(chan weather.Forecast)
	airChan := make(chan airResult, 1)
	errChan := make(chan error, 2)

	// Fetch weather forecast
	go func() {
		forecast, err := weather.GetNOAAWeatherForecast(coords)
		if err != nil {
			errChan <- fmt.Errorf("failed to retrieve weather forecast from NOAA: %w", err)
			return
		}
		weatherChan <- forecast
	}()

	// Fetch air quality forecast
	go func() {
		if config.AirNowAPIKey != "" && coords.Zip != "" {
			anURL := air.BuildAirNowURL(air.AirNowAddress, coords.Zip, config.AirNowAPIKey)
			forecast, err := air.GetForecast(anURL)
			var note string
			if err != nil {
				note = fmt.Sprintf("Air quality unavailable (AirNow: %v)", err)
			}
			airChan <- airResult{forecast, note}
		} else {
			airChan <- airResult{forecast: []air.Forecast{}}
		}
	}()

	// Wait for results with timeout
	select {
	case err := <-errChan:
		return weather.Forecast{}, nil, err
	case wf := <-weatherChan:
		ar := <-airChan
		if ar.note != "" {
			notes = append(notes, ar.note)
		}
		return wf, ar.forecast, nil
	case <-time.After(30 * time.Second):
		return weather.Forecast{}, nil, fmt.Errorf("timeout: weather service did not respond within 30 seconds")
	}
}

// targetZip returns the zip code this run asks for: the -zip flag, else the saved default.
// Empty means IP-based location (no default saved, or -current).
func targetZip(appConfig storage.AppConfig) string {
	if useCurrentLocation {
		return ""
	}
	if zipCode != "" {
		return zipCode
	}
	return appConfig.Config.DefaultZipCode
}

// staleCacheLimit is the oldest cache shown when fetching fails. Older than this,
// most of the hourly forecast has already passed.
const staleCacheLimit = 24 * time.Hour

// loadCachedForecasts returns forecasts from the cache when they are fresh and for the
// location this run asks for. -refresh and -current always bypass the cache; a zip
// request only uses a cache saved for that same zip.
// With stale set (fetching failed), an expired cache up to staleCacheLimit old is accepted,
// even after -refresh; -current still never uses it, since the cache may be for elsewhere.
func loadCachedForecasts(appConfig storage.AppConfig, stale bool) (storage.APICallInfo, weather.Forecast, []air.Forecast, bool) {
	var none storage.APICallInfo
	if useCurrentLocation || (refresh && !stale) {
		return none, weather.Forecast{}, nil, false
	}

	pc, err := storage.LoadCallInfo(appConfig.HomeDir + storage.SavedCallFileName)
	if err != nil {
		return none, weather.Forecast{}, nil, false
	}
	if fresh := isValid(pc.Time, appConfig.CacheTimeoutMinutes); !fresh && (!stale || time.Since(pc.Time) > staleCacheLimit) {
		return none, weather.Forecast{}, nil, false
	}
	if zip := targetZip(appConfig); zip != "" && pc.Coordinates.Zip != zip {
		return none, weather.Forecast{}, nil, false
	}

	pwf, err := storage.LoadSavedWeather(appConfig.HomeDir + storage.SavedWeatherFileName)
	if err != nil {
		return none, weather.Forecast{}, nil, false
	}

	paf, err := storage.LoadSavedAir(appConfig.HomeDir + storage.SavedAirFileName)
	if err != nil {
		notes = append(notes, "Air quality unavailable (no cached air forecast)")
		paf = []air.Forecast{}
	}

	return pc, pwf, paf, true
}

// render prints the header, the selected report, and any notes.
// A zero cachedAt means the forecast was just fetched; stale marks an expired cache.
func render(t time.Time, c geolocation.Coordinates, cachedAt time.Time, stale bool, wf weather.Forecast, af []air.Forecast) {
	PrintHeader(t, c, cachedAt, stale, locatedByIP)
	selected.run(wf, af)
	report.TW.Flush()
	PrintNotes()
}

// saveDefaultZip remembers a -zip request as the default location for later runs,
// and says so, since a lookup quietly changing later runs would be a surprise.
// -zip=ip clears the default instead.
func saveDefaultZip(appConfig storage.AppConfig) {
	if forgetZip && appConfig.Config.DefaultZipCode != "" {
		if err := storage.UpdateDefaultZipCode(appConfig.HomeDir, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Note: Could not clear default zip code: %v\n", err)
			return
		}
		fmt.Fprintln(os.Stderr, "Note: default location cleared; vaporwair now uses IP location")
		return
	}
	if zipCode == "" || zipCode == appConfig.Config.DefaultZipCode {
		return
	}
	if err := storage.UpdateDefaultZipCode(appConfig.HomeDir, zipCode); err != nil {
		fmt.Fprintf(os.Stderr, "Note: Could not save default zip code: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "Note: %s is now your default location (-zip=ip clears it)\n", zipCode)
}

// selectReport parses args and returns the chosen report. The report name may come
// before or after the options; with none, it is the default. "help" returns flag.ErrHelp,
// like -help.
func selectReport(fs *flag.FlagSet, args []string) (reportSpec, error) {
	if err := fs.Parse(args); err != nil {
		return reportSpec{}, err
	}
	spec := reports[0]
	if fs.NArg() > 0 {
		name := fs.Arg(0)
		if name == "help" {
			return reportSpec{}, flag.ErrHelp
		}
		i := slices.IndexFunc(reports, func(r reportSpec) bool { return r.name == name })
		if i < 0 {
			return reportSpec{}, fmt.Errorf("unknown report %q (reports: %s)", name, reportNames())
		}
		spec = reports[i]
		if err := fs.Parse(fs.Args()[1:]); err != nil {
			return reportSpec{}, err
		}
		if fs.NArg() > 0 {
			return reportSpec{}, fmt.Errorf("choose one report; unexpected argument %q", fs.Arg(0))
		}
	}
	if strings.EqualFold(zipCode, ipZip) {
		zipCode, useCurrentLocation, forgetZip = "", true, true
	}
	return spec, nil
}

// reportNames lists the report names for error messages.
func reportNames() string {
	names := make([]string, len(reports))
	for i, r := range reports {
		names[i] = r.name
	}
	return strings.Join(names, ", ")
}

// main orchestrates the application flow: setup, caching, fetching, and reporting.
func main() {
	t := time.Now()
	log.SetFlags(0) // errors read as CLI messages, not timestamped log lines
	fs := newFlagSet()
	spec, err := selectReport(fs, os.Args[1:])
	if errors.Is(err, flag.ErrHelp) {
		printUsage(fs)
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "vaporwair: %v\nRun 'vaporwair help' for usage.\n", err)
		os.Exit(2)
	}
	selected = spec

	// Setup configuration first (may prompt for user input)
	appConfig, err := setupConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	// Serve from cache when it is fresh and for the requested location
	if pc, wf, af, ok := loadCachedForecasts(appConfig, false); ok {
		locatedByIP = targetZip(appConfig) == ""
		render(t, pc.Coordinates, pc.Time, false, wf, af)
		saveDefaultZip(appConfig)
		return
	}

	// Cache miss or expired - fetch new forecasts
	startSpinner("Fetching forecast...")
	coordinates, err := GetCoordinates(appConfig)
	var wf weather.Forecast
	var af []air.Forecast
	if err == nil {
		wf, af, err = fetchForecasts(coordinates, appConfig.Config)
	}
	stopSpinner()
	if err != nil {
		// An older forecast beats no forecast; the header marks it offline.
		if pc, wf, af, ok := loadCachedForecasts(appConfig, true); ok {
			notes = append(notes, fmt.Sprintf("Showing the last saved forecast: %v", err))
			locatedByIP = targetZip(appConfig) == ""
			render(t, pc.Coordinates, pc.Time, true, wf, af)
			return
		}
		fatalf("%v", err)
	}

	render(t, coordinates, time.Time{}, false, wf, af)
	SaveForecasts(appConfig.HomeDir, coordinates, wf, af)
	saveDefaultZip(appConfig)
}
