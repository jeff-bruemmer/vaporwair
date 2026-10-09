package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/dialer"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"github.com/jeff-bruemmer/vaporwair/src/report"
	"github.com/jeff-bruemmer/vaporwair/src/storage"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// version is printed by -version and sent in the User-Agent.
const version = "3.0.0"

// docsURL is where the docs live and issues go.
const docsURL = "https://github.com/jeff-bruemmer/vaporwair"

// reportSpec is one report: its subcommand, a one-line description for usage,
// how it prints, and its data for -json.
type reportSpec struct {
	name, desc string
	run        func(weather.Forecast, []air.Forecast)
	data       func(weather.Forecast, []air.Forecast) map[string]any
}

// reports lists every report; the first is the default.
var reports = []reportSpec{
	{"insights", "Today's forecast, what to wear, and the next few hours", report.InsightsReport, report.InsightsData},
	{"summary", "A few lines: now, high and low, outfit, air, and alerts", report.Summary, report.SummaryData},
	{"hourly", "Hour by hour for the next 12 hours", report.WeatherHourly, report.HourlyData},
	{"week", "One line per day for the next 7 days", report.WeatherWeek, report.DaysData},
	{"daily", "Every NOAA field for each of the next 7 days", report.WeatherDaily, report.DaysData},
	{"alerts", "Active weather alerts and warnings", report.WeatherAlertsReport, report.AlertsData},
	{"air", "Air quality forecast by pollutant", report.AirQuality, report.AirData},
	{"clothing", "What to wair, with the full outfit scale", report.ClothingReport, report.ClothingData},
}

// selected is the report this run prints.
var selected = reports[0]

// Location, cache, and output options
var zipCode string
var defaultZip string
var useCurrentLocation bool
var refresh bool
var jsonOutput bool
var showVersion bool

// clearDefault is set by -default=ip: clear the saved default and use IP-based location.
var clearDefault bool

// ipZip is the -default value that clears the saved default.
const ipZip = "ip"

// options are the flags shown in usage, in order.
var options = []string{"zip", "default", "current", "refresh", "json", "version"}

// notes collects non-fatal warnings raised while fetching.
// They're printed to stderr after the report so they don't garble the spinner line.
// Only the main goroutine appends to it.
var notes []string

// stopSpinner clears the spinner if one is running. It is safe to call more than once,
// from more than one goroutine.
var stopSpinner = func() {}

// hintError is an error with a suggestion printed on the line below it. The hint stays
// out of Error(), so a note that quotes the error stays on one line.
type hintError struct {
	error
	hint string
}

func (h hintError) Unwrap() error { return h.error }

// fatal stops the spinner, prints err (and any hint) to stderr, and exits.
func fatal(err error) {
	stopSpinner()
	fmt.Fprintf(os.Stderr, "vaporwair: %v\n", err)
	var h hintError
	if errors.As(err, &h) {
		fmt.Fprintln(os.Stderr, h.hint)
	}
	os.Exit(1)
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
				// A few extra spaces also cover the "^C" the terminal echoes on Ctrl-C.
				fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", len(label)+4))
				return
			case <-tick.C:
			}
		}
	}()
	var once sync.Once
	stopSpinner = func() {
		once.Do(func() {
			close(done)
			<-finished
		})
	}
}

// exitOnInterrupt makes Ctrl-C clear the spinner and exit at once, rather than leave a
// half-drawn spinner line behind. The returned function restores the default handling.
func exitOnInterrupt(stop func()) func() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		if _, ok := <-sig; ok {
			stop()
			os.Exit(130) // 128 + SIGINT, as a shell reports it
		}
	}()
	return func() {
		signal.Stop(sig)
		close(sig)
	}
}

// locatedByIP is true when this run's location came from IP lookup rather than a zip code.
// The header says so, since an IP can place you far from where you are.
var locatedByIP bool

// getIPGeoData retrieves geolocation data from IP address.
func getIPGeoData() (geolocation.Coordinates, error) {
	geoData, err := geolocation.GetGeoData(geolocation.IPAPIAddress)
	if err != nil {
		return geolocation.Coordinates{}, hintError{fmt.Errorf("locating by IP address: %w", err), "Try a zip code instead: vaporwair -zip CODE"}
	}
	locatedByIP = true
	return geolocation.FormatCoordinates(geoData), nil
}

// zipCoordinates returns a zip code's coordinates. When the last forecast was fetched for
// the same zip, it reuses the coordinates saved with it rather than looking them up again.
func zipCoordinates(appConfig storage.AppConfig, zip string) (geolocation.Coordinates, error) {
	pc, err := storage.LoadCallInfo(appConfig.CacheFile(storage.SavedCallFileName))
	if err == nil && !pc.ByIP && pc.Coordinates.Zip == zip && pc.Coordinates.Latitude != "" && pc.Coordinates.Longitude != "" {
		return pc.Coordinates, nil
	}
	geoData, err := geolocation.GetGeoDataFromZip(zip)
	if err != nil {
		return geolocation.Coordinates{}, err
	}
	return geolocation.FormatCoordinates(geoData), nil
}

// GetCoordinates retrieves coordinates based on zip code flag, default zip, or IP address.
// Priority: 1) -current flag (IP), 2) -zip flag, 3) default zip (-default or config), 4) IP geolocation
func GetCoordinates(appConfig storage.AppConfig) (geolocation.Coordinates, error) {
	if useCurrentLocation {
		return getIPGeoData()
	}

	if zipCode != "" {
		return zipCoordinates(appConfig, zipCode)
	}

	// A saved default that fails to look up falls back to IP. One just given with -default
	// doesn't: it would be saved after a run that never used it.
	if def := appConfig.Config.DefaultZipCode; def != "" {
		coords, err := zipCoordinates(appConfig, def)
		if err == nil || defaultZip != "" {
			return coords, err
		}
		coords, ipErr := getIPGeoData()
		if ipErr != nil {
			return coords, err
		}
		notes = append(notes, fmt.Sprintf("Could not use default zip code %s (%v); used IP-based location", def, err))
		return coords, nil
	}

	return getIPGeoData()
}

// printBanner prints the name banner that heads the usage text.
func printBanner(w io.Writer) {
	fmt.Fprint(w, "\nV A P O R W A I R\n\n")
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

// PrintNotes prints any warnings collected while fetching to stderr, after the report,
// so they never end up in piped or redirected output.
func PrintNotes() {
	if len(notes) > 0 && !jsonOutput {
		fmt.Fprintln(os.Stderr)
	}
	for _, n := range notes {
		fmt.Fprintln(os.Stderr, "Note: "+n)
	}
}

// newFlagSet defines the options. The flag package's own error and usage output is off:
// main prints a short error, or the full usage for -help.
func newFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("vaporwair", flag.ContinueOnError)
	fs.StringVar(&zipCode, "zip", "", "Weather for a US zip code, this once")
	fs.StringVar(&defaultZip, "default", "", "Make a zip code your default location (-default=ip clears it)")
	fs.BoolVar(&useCurrentLocation, "current", false, "Use IP-based location this once")
	fs.BoolVar(&refresh, "refresh", false, "Skip the 5-minute cache and fetch fresh forecasts")
	fs.BoolVar(&jsonOutput, "json", false, "Print the report's data as JSON")
	fs.BoolVar(&showVersion, "version", false, "Print the version")
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	return fs
}

// printUsage prints the help text: what vaporwair is, examples, reports, and options.
func printUsage(w io.Writer, fs *flag.FlagSet) {
	printBanner(w)
	fmt.Fprintln(w, "Weather and air quality reports for US locations, from NOAA and EPA AirNow.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage: vaporwair [report] [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	for _, ex := range [][2]string{
		{"vaporwair", "Today's forecast for your location"},
		{"vaporwair hourly -zip=10001", "The next 12 hours in New York, this once"},
		{"vaporwair -default=05401", "Make Burlington, VT your default location"},
		{"vaporwair summary -json", "The summary as JSON, for scripts"},
	} {
		fmt.Fprintf(w, "  %-28s %s\n", ex[0], ex[1])
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Reports:")
	for i, r := range reports {
		desc := r.desc
		if i == 0 {
			desc += " (default)"
		}
		fmt.Fprintf(w, "  %-10s %s\n", r.name, desc)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	for _, name := range options {
		f := fs.Lookup(name)
		arg := ""
		if name == "zip" || name == "default" {
			arg = " CODE"
		}
		fmt.Fprintf(w, "  %-14s %s\n", "-"+name+arg, f.Usage)
	}
	fmt.Fprintln(w)
	if dir, err := storage.ConfigDir(); err == nil {
		fmt.Fprintf(w, "Config: %s (AirNow API key, default zip)\n", storage.Tilde(filepath.Join(dir, storage.ConfigFileName)))
	}
	fmt.Fprintf(w, "Docs and issues: %s\n", docsURL)
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
		configFile := storage.Tilde(appConfig.ConfigFile())
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
			errChan <- err
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
				note = fmt.Sprintf("Air quality unavailable: %v", err)
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
		return weather.Forecast{}, nil, errors.New("the weather service did not respond within 30 seconds")
	}
}

// targetZip returns the zip code this run asks for: the -zip flag, else the default.
// Empty means IP-based location (no default, or -current).
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
// request only uses a cache saved for that same zip, and an IP-located run only one
// that was itself located by IP.
// With stale set (fetching failed), an expired cache up to staleCacheLimit old is accepted,
// even after -refresh; -current still never uses it, since the cache may be for elsewhere.
func loadCachedForecasts(appConfig storage.AppConfig, stale bool) (storage.APICallInfo, weather.Forecast, []air.Forecast, bool) {
	var none storage.APICallInfo
	if useCurrentLocation || (refresh && !stale) {
		return none, weather.Forecast{}, nil, false
	}

	pc, err := storage.LoadCallInfo(appConfig.CacheFile(storage.SavedCallFileName))
	if err != nil {
		return none, weather.Forecast{}, nil, false
	}
	if fresh := isValid(pc.Time, appConfig.CacheTimeoutMinutes); !fresh && (!stale || time.Since(pc.Time) > staleCacheLimit) {
		return none, weather.Forecast{}, nil, false
	}
	if zip := targetZip(appConfig); zip == "" && !pc.ByIP || zip != "" && (pc.ByIP || pc.Coordinates.Zip != zip) {
		return none, weather.Forecast{}, nil, false
	}

	pwf, err := storage.LoadSavedWeather(appConfig.CacheFile(storage.SavedWeatherFileName))
	if err != nil {
		return none, weather.Forecast{}, nil, false
	}

	paf, err := storage.LoadSavedAir(appConfig.CacheFile(storage.SavedAirFileName))
	if err != nil {
		notes = append(notes, "Air quality unavailable (no cached air forecast)")
		paf = []air.Forecast{}
	}

	return pc, pwf, paf, true
}

// render prints the header, the selected report, and any notes; with -json, the report's
// data instead of its text. A zero cachedAt means the forecast was just fetched; stale
// marks an expired cache.
func render(t time.Time, c geolocation.Coordinates, cachedAt time.Time, stale bool, wf weather.Forecast, af []air.Forecast) {
	wf, af = report.DropPast(wf, af, t)
	if jsonOutput {
		printJSON(t, c, cachedAt, stale, wf, af)
	} else {
		PrintHeader(t, c, cachedAt, stale, locatedByIP)
		selected.run(wf, af)
		report.TW.Flush()
	}
	PrintNotes()
}

// locationJSON is the "location" object in -json output.
type locationJSON struct {
	City      string  `json:"city"`
	Zip       string  `json:"zip"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ByIP      bool    `json:"by_ip"`
}

// printJSON prints the selected report's data with what the header would say:
// where, when, and whether it came from the cache.
func printJSON(t time.Time, c geolocation.Coordinates, cachedAt time.Time, stale bool, wf weather.Forecast, af []air.Forecast) {
	data := selected.data(wf, af)
	data["report"] = selected.name
	data["generated_at"] = t.Format(time.RFC3339)
	data["cached_at"] = nil
	if !cachedAt.IsZero() {
		data["cached_at"] = cachedAt.Format(time.RFC3339)
	}
	data["offline"] = stale
	lat, _ := strconv.ParseFloat(c.Latitude, 64)
	lon, _ := strconv.ParseFloat(c.Longitude, 64)
	data["location"] = locationJSON{City: c.City, Zip: c.Zip, Latitude: lat, Longitude: lon, ByIP: locatedByIP}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fatal(err)
	}
	os.Stdout.Write(append(out, '\n'))
}

// applyDefaultFlag puts -default into this run's config, so the location priority
// (-current, -zip, default, IP) treats it like a saved default.
func applyDefaultFlag(appConfig *storage.AppConfig) {
	switch {
	case clearDefault:
		appConfig.Config.DefaultZipCode = ""
		// The cache may hold the old default's forecast; fetch the IP location instead.
		if zipCode == "" {
			useCurrentLocation = true
		}
	case defaultZip != "":
		appConfig.Config.DefaultZipCode = defaultZip
	}
}

// saveDefaultZip saves a -default request for later runs and says so, since a change to
// later runs should never be a surprise. saved is the default before this run.
func saveDefaultZip(configFile, saved string) {
	want := saved
	switch {
	case clearDefault:
		want = ""
	case defaultZip != "":
		want = defaultZip
	}
	if want == saved {
		return
	}
	if err := storage.UpdateDefaultZipCode(configFile, want); err != nil {
		fmt.Fprintf(os.Stderr, "Note: could not save the default zip code: %v\n", err)
		return
	}
	if want == "" {
		fmt.Fprintln(os.Stderr, "Note: default location cleared; vaporwair now uses IP location")
		return
	}
	fmt.Fprintf(os.Stderr, "Note: %s is now your default location (-default=ip clears it)\n", want)
}

// selectReport parses and checks args, and returns the chosen report. The report name may
// come before or after the options; with none, it is the default. "help" returns
// flag.ErrHelp, like -help.
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
			if guess := suggestReport(name); guess != "" {
				return reportSpec{}, fmt.Errorf("unknown report %q; did you mean %q?", name, guess)
			}
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
	return spec, checkLocationFlags(fs)
}

// checkLocationFlags validates -zip and -default before anything is fetched or saved.
func checkLocationFlags(fs *flag.FlagSet) error {
	given := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { given[f.Name] = true })

	if given["zip"] {
		if strings.EqualFold(zipCode, ipZip) {
			return errors.New("-zip takes a zip code; use -current for IP location, or -default=ip to clear your default")
		}
		if err := geolocation.ValidateZip(zipCode); err != nil {
			return fmt.Errorf("-zip: %w", err)
		}
		if useCurrentLocation {
			return errors.New("choose one location: -zip or -current")
		}
	}
	if given["default"] {
		if strings.EqualFold(defaultZip, ipZip) {
			defaultZip, clearDefault = "", true
		} else if err := geolocation.ValidateZip(defaultZip); err != nil {
			return fmt.Errorf("-default: %w", err)
		}
	}
	return nil
}

// suggestReport returns the report name most like name, or "" when none is close:
// one within two edits ("hourl", "weekly"), or else the only one sharing name's first
// four letters ("clothes"). The suggestion is shown, never run.
func suggestReport(name string) string {
	name = strings.ToLower(name)
	best, bestDist := "", 3
	for _, r := range reports {
		if d := editDistance(name, r.name); d < bestDist {
			best, bestDist = r.name, d
		}
	}
	if best != "" || len(name) < 4 {
		return best
	}
	var matches []string
	for _, r := range reports {
		if strings.HasPrefix(r.name, name[:4]) {
			matches = append(matches, r.name)
		}
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return ""
}

// editDistance is the Levenshtein distance between a and b.
func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
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
	dialer.UserAgent = "vaporwair/" + version + " (" + docsURL + ")"
	fs := newFlagSet()
	spec, err := selectReport(fs, os.Args[1:])
	if errors.Is(err, flag.ErrHelp) {
		printUsage(os.Stdout, fs)
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "vaporwair: %v\nRun 'vaporwair help' for usage.\n", err)
		os.Exit(2)
	}
	if showVersion {
		fmt.Println("vaporwair " + version)
		return
	}
	selected = spec

	// Setup configuration first (may prompt for user input)
	appConfig, err := setupConfiguration()
	if err != nil {
		fatal(err)
	}
	savedDefault := appConfig.Config.DefaultZipCode
	applyDefaultFlag(&appConfig)

	// Serve from cache when it is fresh and for the requested location
	if pc, wf, af, ok := loadCachedForecasts(appConfig, false); ok {
		locatedByIP = pc.ByIP
		render(t, pc.Coordinates, pc.Time, false, wf, af)
		saveDefaultZip(appConfig.ConfigFile(), savedDefault)
		return
	}

	// Cache miss or expired - fetch new forecasts
	startSpinner("Fetching forecast...")
	restoreInterrupt := exitOnInterrupt(stopSpinner)
	coordinates, err := GetCoordinates(appConfig)
	var wf weather.Forecast
	var af []air.Forecast
	if err == nil {
		wf, af, err = fetchForecasts(coordinates, appConfig.Config)
	}
	stopSpinner()
	restoreInterrupt()
	if err != nil {
		// An older forecast beats no forecast; the header marks it offline.
		if pc, wf, af, ok := loadCachedForecasts(appConfig, true); ok {
			notes = append(notes, fmt.Sprintf("Showing the last saved forecast: %v", err))
			locatedByIP = pc.ByIP
			render(t, pc.Coordinates, pc.Time, true, wf, af)
			saveDefaultZip(appConfig.ConfigFile(), savedDefault)
			return
		}
		fatal(err)
	}

	render(t, coordinates, time.Time{}, false, wf, af)
	if err := storage.SaveForecasts(appConfig, coordinates, locatedByIP, wf, af); err != nil {
		fmt.Fprintf(os.Stderr, "Note: could not save the forecast cache: %v\n", err)
	}
	saveDefaultZip(appConfig.ConfigFile(), savedDefault)
}
