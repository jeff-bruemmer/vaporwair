package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"github.com/jeff-bruemmer/vaporwair/src/storage"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// cacheAged writes a cache for zip 05401 saved age ago into a temp cache directory.
// byIP marks it as located by IP; defaultZip is the saved default in the returned config.
func cacheAged(t *testing.T, age time.Duration, byIP bool, defaultZip string) storage.AppConfig {
	t.Helper()
	a := storage.AppConfig{ConfigDir: t.TempDir(), CacheDir: t.TempDir(), CacheTimeoutMinutes: 5}
	a.Config.DefaultZipCode = defaultZip
	coords := geolocation.Coordinates{Latitude: "44.4759", Longitude: "-73.2121", City: "Burlington", Zip: "05401"}
	if err := storage.SaveForecasts(a, coords, byIP, weather.Forecast{Timezone: "America/New_York"}, []air.Forecast{}); err != nil {
		t.Fatal(err)
	}
	info := storage.APICallInfo{Time: time.Now().Add(-age), Coordinates: coords, ByIP: byIP}
	if err := storage.SaveCall(a.CacheFile(storage.SavedCallFileName), info); err != nil {
		t.Fatal(err)
	}
	return a
}

// setFlags sets the location flags for one test and restores every option afterwards.
func setFlags(t *testing.T, zip string, current, refreshFlag bool) {
	t.Helper()
	oldZip, oldDefault, oldClear, oldCurrent, oldRefresh := zipCode, defaultZip, clearDefault, useCurrentLocation, refresh
	oldJSON, oldVersion, oldNotes, oldByIP, oldSelected := jsonOutput, showVersion, notes, locatedByIP, selected
	zipCode, useCurrentLocation, refresh = zip, current, refreshFlag
	defaultZip, clearDefault, jsonOutput, showVersion, notes, locatedByIP = "", false, false, false, nil, false
	t.Cleanup(func() {
		zipCode, defaultZip, clearDefault, useCurrentLocation, refresh = oldZip, oldDefault, oldClear, oldCurrent, oldRefresh
		jsonOutput, showVersion, notes, locatedByIP, selected = oldJSON, oldVersion, oldNotes, oldByIP, oldSelected
	})
}

func TestLoadCachedForecasts(t *testing.T) {
	tests := []struct {
		name       string
		age        time.Duration
		byIP       bool
		defaultZip string
		zip        string
		current    bool
		refresh    bool
		stale      bool
		want       bool
	}{
		{"fresh cache is used", time.Minute, false, "05401", "", false, false, false, true},
		{"expired cache is not used normally", 2 * time.Hour, false, "05401", "", false, false, false, false},
		{"expired cache is used when offline", 2 * time.Hour, false, "05401", "", false, false, true, true},
		{"offline fallback works after -refresh", 2 * time.Hour, false, "05401", "", false, true, true, true},
		{"cache older than a day is not used offline", 25 * time.Hour, false, "05401", "", false, false, true, false},
		{"offline fallback respects the requested zip", 2 * time.Hour, false, "05401", "10001", false, false, true, false},
		{"-current never uses the cache", 2 * time.Hour, false, "05401", "", true, false, true, false},
		{"-refresh bypasses a fresh cache", time.Minute, false, "05401", "", false, true, false, false},
		{"a -zip cache is used for the same -zip", time.Minute, false, "", "05401", false, false, false, true},
		{"a one-off -zip cache is not shown as the IP location", time.Minute, false, "", "", false, false, false, false},
		{"an IP cache is used for IP location", time.Minute, true, "", "", false, false, false, true},
		{"an IP cache is not used for a zip", time.Minute, true, "", "05401", false, false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appConfig := cacheAged(t, tt.age, tt.byIP, tt.defaultZip)
			setFlags(t, tt.zip, tt.current, tt.refresh)
			if _, _, _, got := loadCachedForecasts(appConfig, tt.stale); got != tt.want {
				t.Errorf("loadCachedForecasts(stale=%v) = %v, want %v", tt.stale, got, tt.want)
			}
		})
	}
}

// A missing cache file is a cache miss, not a fatal error.
func TestLoadCachedForecastsMissingWeatherFile(t *testing.T) {
	appConfig := cacheAged(t, time.Minute, false, "05401")
	setFlags(t, "", false, false)
	os.Remove(appConfig.CacheFile(storage.SavedWeatherFileName))
	if _, _, _, ok := loadCachedForecasts(appConfig, false); ok {
		t.Error("expected a cache miss when the weather file is missing")
	}
}

// A zip code's saved coordinates are reused, so most runs skip the zip lookup service,
// but never coordinates found by IP lookup.
func TestZipCoordinatesReusesCache(t *testing.T) {
	appConfig := cacheAged(t, 2*time.Hour, false, "05401")
	got, err := zipCoordinates(appConfig, "05401")
	if err != nil || got.Latitude != "44.4759" || got.City != "Burlington" {
		t.Errorf("zipCoordinates = %+v, %v; want the cached Burlington coordinates", got, err)
	}

	// An IP-located cache must not stand in for a zip lookup. "0540x" fails validation
	// before any request, so reaching the lookup shows the cache was skipped.
	ipConfig := storage.AppConfig{CacheDir: t.TempDir()}
	info := storage.APICallInfo{Time: time.Now(), Coordinates: geolocation.Coordinates{Latitude: "1", Longitude: "2", Zip: "0540x"}, ByIP: true}
	if err := storage.SaveCall(ipConfig.CacheFile(storage.SavedCallFileName), info); err != nil {
		t.Fatal(err)
	}
	if _, err := zipCoordinates(ipConfig, "0540x"); err == nil {
		t.Error("zipCoordinates reused IP-located coordinates for a zip code")
	}
}

func TestPrintHeaderMarksOfflineCache(t *testing.T) {
	now := time.Date(2026, time.October, 6, 19, 56, 0, 0, time.Local)
	c := geolocation.Coordinates{City: "Burlington", Zip: "05401"}
	tests := []struct {
		cachedAt time.Time
		stale    bool
		byIP     bool
		want     string
	}{
		{time.Time{}, false, false, "Burlington 05401 | Tue Oct 6, 19:56"},
		{time.Time{}, false, true, "Burlington 05401 (via IP) | Tue Oct 6, 19:56"},
		{now.Add(-3 * time.Minute), false, false, "| cached 3m ago\n"},
		{now.Add(-2 * time.Hour), true, false, "| cached 2h ago (offline)\n"},
	}
	for _, tt := range tests {
		out := captureOutput(t, &os.Stdout, func() { PrintHeader(now, c, tt.cachedAt, tt.stale, tt.byIP) })
		if !strings.Contains(out, tt.want) {
			t.Errorf("PrintHeader = %q, want it to contain %q", out, tt.want)
		}
	}
}

// Warnings go to stderr, so piped or redirected reports stay clean.
func TestNotesGoToStderr(t *testing.T) {
	setFlags(t, "", false, false)
	notes = []string{"Air quality unavailable"}
	var stderr string
	stdout := captureOutput(t, &os.Stdout, func() {
		stderr = captureOutput(t, &os.Stderr, PrintNotes)
	})
	if stdout != "" {
		t.Errorf("notes on stdout: %q", stdout)
	}
	if !strings.Contains(stderr, "Note: Air quality unavailable") {
		t.Errorf("notes missing from stderr: %q", stderr)
	}
}

// captureOutput returns what run writes to *f (os.Stdout or os.Stderr).
func captureOutput(t *testing.T, f **os.File, run func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := *f
	*f = w
	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()
	run()
	*f = old
	w.Close()
	return <-done
}

func TestSelectReport(t *testing.T) {
	cases := []struct {
		args    []string
		want    string
		wantZip string
		wantErr error // nil, flag.ErrHelp, or errAny
	}{
		{nil, "insights", "", nil},
		{[]string{"hourly"}, "hourly", "", nil},
		{[]string{"-zip=05401", "week"}, "week", "05401", nil},
		{[]string{"week", "-zip=05401"}, "week", "05401", nil},
		{[]string{"week", "--zip", "05401"}, "week", "05401", nil},
		{[]string{"-h"}, "", "", flag.ErrHelp},
		{[]string{"-help"}, "", "", flag.ErrHelp},
		{[]string{"hourly", "-h"}, "", "", flag.ErrHelp},
		{[]string{"help"}, "", "", flag.ErrHelp},
		{[]string{"-zip=05401", "help"}, "", "", flag.ErrHelp},
		{[]string{"-w"}, "", "", errAny}, // report flags are gone; reports are named
		{[]string{"hourly", "week"}, "", "", errAny},
		{[]string{"bogus"}, "", "", errAny},
		{[]string{"-zip=1234"}, "", "", errAny},
		{[]string{"-zip=0540a"}, "", "", errAny},
		{[]string{"-zip=05401-1234"}, "", "", errAny},
		{[]string{"-zip="}, "", "", errAny},
		{[]string{"-zip=ip"}, "", "", errAny},
		{[]string{"-zip=05401", "-current"}, "", "", errAny},
		{[]string{"-default=123"}, "", "", errAny},
		{[]string{"-default=05401", "-zip=10001"}, "insights", "10001", nil},
	}
	for _, c := range cases {
		setFlags(t, "", false, false)
		got, err := selectReport(newFlagSet(), c.args)
		switch {
		case c.wantErr == errAny && err != nil:
		case c.wantErr != nil && errors.Is(err, c.wantErr):
		case c.wantErr == nil && err == nil:
			if got.name != c.want || zipCode != c.wantZip {
				t.Errorf("%v: got %q zip %q, want %q zip %q", c.args, got.name, zipCode, c.want, c.wantZip)
			}
		default:
			t.Errorf("%v: err = %v, want %v", c.args, err, c.wantErr)
		}
	}
}

// errAny stands for "some error" in TestSelectReport.
var errAny = errors.New("any error")

// Errors that name a fix say so: -zip=ip points to the flags that do what it used to mean.
func TestSelectReportErrorHints(t *testing.T) {
	cases := map[string]string{
		"-zip=ip":         "-default=ip",
		"-zip=05401-1234": "use the 5-digit form",
		"weekly":          `did you mean "week"?`,
	}
	for arg, want := range cases {
		setFlags(t, "", false, false)
		_, err := selectReport(newFlagSet(), []string{arg})
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err = %v, want it to contain %q", arg, err, want)
		}
	}
}

// -default=ip clears the saved default and uses IP location for this run, so a cache for
// the old default is not shown as the IP location.
func TestDefaultIPClearsDefault(t *testing.T) {
	setFlags(t, "", false, false)
	if _, err := selectReport(newFlagSet(), []string{"-default=IP"}); err != nil {
		t.Fatal(err)
	}
	if !clearDefault || defaultZip != "" {
		t.Fatalf("-default=IP: clear %v zip %q, want true \"\"", clearDefault, defaultZip)
	}
	a := storage.AppConfig{Config: storage.Config{DefaultZipCode: "05401"}}
	applyDefaultFlag(&a)
	if a.Config.DefaultZipCode != "" || !useCurrentLocation {
		t.Errorf("after -default=ip: default %q current %v", a.Config.DefaultZipCode, useCurrentLocation)
	}
}

// -zip is a one-off; only -default changes the saved default, and says so.
func TestSaveDefaultZip(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), storage.ConfigFileName)
	if err := storage.CreateConfig(configFile, "key"); err != nil {
		t.Fatal(err)
	}
	saved := func() string {
		c, err := storage.GetConfig(configFile)
		if err != nil {
			t.Fatal(err)
		}
		return c.DefaultZipCode
	}

	setFlags(t, "10001", false, false)
	if out := captureOutput(t, &os.Stderr, func() { saveDefaultZip(configFile, "") }); out != "" || saved() != "" {
		t.Errorf("-zip changed the default to %q (stderr %q)", saved(), out)
	}

	setFlags(t, "", false, false)
	defaultZip = "05401"
	out := captureOutput(t, &os.Stderr, func() { saveDefaultZip(configFile, "") })
	if saved() != "05401" || !strings.Contains(out, "05401 is now your default location") {
		t.Errorf("-default=05401: saved %q, stderr %q", saved(), out)
	}

	defaultZip, clearDefault = "", true
	out = captureOutput(t, &os.Stderr, func() { saveDefaultZip(configFile, "05401") })
	if saved() != "" || !strings.Contains(out, "default location cleared") {
		t.Errorf("-default=ip: saved %q, stderr %q", saved(), out)
	}
}

func TestSuggestReport(t *testing.T) {
	cases := map[string]string{
		"hourl":    "hourly",
		"Hourly":   "hourly",
		"weekly":   "week",
		"alert":    "alerts",
		"clothes":  "clothing",
		"insight":  "insights",
		"bogus":    "",
		"forecast": "",
	}
	for in, want := range cases {
		if got := suggestReport(in); got != want {
			t.Errorf("suggestReport(%q) = %q, want %q", in, got, want)
		}
	}
}

// Requested help goes to stdout, lists every report and option, and fits 80 columns.
func TestUsage(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf, newFlagSet())
	out := buf.String()
	var wants []string
	for _, r := range reports {
		wants = append(wants, "  "+r.name+" ")
	}
	for _, name := range options {
		wants = append(wants, "  -"+name)
	}
	wants = append(wants, "Examples:", docsURL, "Config: ")
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("usage missing %q:\n%s", want, out)
		}
	}
	for _, line := range strings.Split(out, "\n") {
		if len(line) > 80 {
			t.Errorf("usage line is %d columns: %q", len(line), line)
		}
	}
}

// Every report's -json output is valid JSON with the envelope and the report's keys.
func TestPrintJSON(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	var w weather.Forecast
	for i := range 24 {
		w.Hourly.Data = append(w.Hourly.Data, weather.DataPoint{Time: float64(now.Add(time.Duration(i) * time.Hour).Unix()), Temperature: 50})
	}
	w.Daily.Data = []weather.DataPoint{{Time: float64(now.Unix()), PeriodName: "Today", TemperatureMax: 60}}
	c := geolocation.Coordinates{Latitude: "44.4759", Longitude: "-73.2121", City: "Burlington", Zip: "05401"}

	reportKeys := map[string][]string{
		"insights": {"now", "today", "outfit", "hours", "alerts", "aqi_today"},
		"summary":  {"now", "today", "outfit", "alerts", "aqi_today"},
		"hourly":   {"hours"},
		"week":     {"days"},
		"daily":    {"days"},
		"alerts":   {"alerts"},
		"air":      {"air"},
		"clothing": {"outfit", "hours"},
	}
	for _, spec := range reports {
		setFlags(t, "", false, false)
		selected, jsonOutput = spec, true
		out := captureOutput(t, &os.Stdout, func() { render(now, c, now.Add(-time.Minute), false, w, nil) })

		var got map[string]any
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("%s -json is not valid JSON: %v\n%s", spec.name, err, out)
		}
		want := append([]string{"report", "generated_at", "cached_at", "offline", "location"}, reportKeys[spec.name]...)
		var keys []string
		for k := range got {
			keys = append(keys, k)
		}
		slices.Sort(want)
		slices.Sort(keys)
		if !slices.Equal(keys, want) {
			t.Errorf("%s -json keys = %v, want %v", spec.name, keys, want)
		}
		if got["report"] != spec.name {
			t.Errorf("%s -json report = %v", spec.name, got["report"])
		}
		if loc, _ := got["location"].(map[string]any); loc["zip"] != "05401" || loc["latitude"] != 44.4759 {
			t.Errorf("%s -json location = %v", spec.name, got["location"])
		}
	}
}
