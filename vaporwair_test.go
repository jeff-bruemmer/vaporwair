package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"github.com/jeff-bruemmer/vaporwair/src/storage"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
)

// cacheAged writes a cache for zip 05401 saved age ago into a temp home directory.
func cacheAged(t *testing.T, age time.Duration) storage.AppConfig {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(home+storage.VaporwairDir, 0755); err != nil {
		t.Fatal(err)
	}
	coords := geolocation.Coordinates{City: "Burlington", Zip: "05401"}
	if err := storage.SaveCall(home+storage.SavedCallFileName, storage.APICallInfo{Time: time.Now().Add(-age), Coordinates: coords}); err != nil {
		t.Fatal(err)
	}
	storage.SaveWeatherForecast(home+storage.SavedWeatherFileName, weather.Forecast{Timezone: "America/New_York"})
	storage.SaveAirForecast(home+storage.SavedAirFileName, []air.Forecast{})
	return storage.AppConfig{HomeDir: home, CacheTimeoutMinutes: 5}
}

// setFlags sets the location flags for one test and restores them afterwards.
func setFlags(t *testing.T, zip string, current, refreshFlag bool) {
	t.Helper()
	oldZip, oldCurrent, oldRefresh, oldNotes := zipCode, useCurrentLocation, refresh, notes
	zipCode, useCurrentLocation, refresh = zip, current, refreshFlag
	t.Cleanup(func() { zipCode, useCurrentLocation, refresh, notes = oldZip, oldCurrent, oldRefresh, oldNotes })
}

func TestLoadCachedForecasts(t *testing.T) {
	tests := []struct {
		name    string
		age     time.Duration
		zip     string
		current bool
		refresh bool
		stale   bool
		want    bool
	}{
		{"fresh cache is used", time.Minute, "", false, false, false, true},
		{"expired cache is not used normally", 2 * time.Hour, "", false, false, false, false},
		{"expired cache is used when offline", 2 * time.Hour, "", false, false, true, true},
		{"offline fallback works after -refresh", 2 * time.Hour, "", false, true, true, true},
		{"cache older than a day is not used offline", 25 * time.Hour, "", false, false, true, false},
		{"offline fallback respects the requested zip", 2 * time.Hour, "10001", false, false, true, false},
		{"-current never uses the cache", 2 * time.Hour, "", true, false, true, false},
		{"-refresh bypasses a fresh cache", time.Minute, "", false, true, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appConfig := cacheAged(t, tt.age)
			setFlags(t, tt.zip, tt.current, tt.refresh)
			if _, _, _, got := loadCachedForecasts(appConfig, tt.stale); got != tt.want {
				t.Errorf("loadCachedForecasts(stale=%v) = %v, want %v", tt.stale, got, tt.want)
			}
		})
	}
}

// A missing cache file is a cache miss, not a fatal error.
func TestLoadCachedForecastsMissingWeatherFile(t *testing.T) {
	appConfig := cacheAged(t, time.Minute)
	setFlags(t, "", false, false)
	os.Remove(appConfig.HomeDir + storage.SavedWeatherFileName)
	if _, _, _, ok := loadCachedForecasts(appConfig, false); ok {
		t.Error("expected a cache miss when the weather file is missing")
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
		out := captureStdout(t, func() { PrintHeader(now, c, tt.cachedAt, tt.stale, tt.byIP) })
		if !strings.Contains(out, tt.want) {
			t.Errorf("PrintHeader = %q, want it to contain %q", out, tt.want)
		}
	}
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()
	run()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
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
		{[]string{"-h"}, "", "", flag.ErrHelp},
		{[]string{"-help"}, "", "", flag.ErrHelp},
		{[]string{"hourly", "-h"}, "", "", flag.ErrHelp},
		{[]string{"help"}, "", "", flag.ErrHelp},
		{[]string{"-zip=05401", "help"}, "", "", flag.ErrHelp},
		{[]string{"-w"}, "week", "", nil}, // old report flags still work, with a warning
		{[]string{"-alerts"}, "alerts", "", nil},
		{[]string{"-c", "-zip=05401"}, "clothing", "05401", nil},
		{[]string{"-w", "-d"}, "", "", errAny},
		{[]string{"hourly", "-w"}, "", "", errAny},
		{[]string{"hourly", "week"}, "", "", errAny},
		{[]string{"bogus"}, "", "", errAny},
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

// -zip=ip uses IP location and clears the saved default; it is not a zip code.
func TestZipIPClearsDefault(t *testing.T) {
	setFlags(t, "", false, false)
	t.Cleanup(func() { forgetZip = false })
	if _, err := selectReport(newFlagSet(), []string{"-zip=ip"}); err != nil {
		t.Fatal(err)
	}
	if zipCode != "" || !useCurrentLocation || !forgetZip {
		t.Errorf("-zip=ip: zip %q current %v forget %v, want \"\" true true", zipCode, useCurrentLocation, forgetZip)
	}
}

// Each old report flag selects its report and names the replacement in its warning.
func TestOldReportFlags(t *testing.T) {
	for name, want := range oldReportFlags {
		setFlags(t, "", false, false)
		fs := newFlagSet()
		got, err := selectReport(fs, []string{"-" + name})
		if err != nil || got.name != want {
			t.Errorf("-%s: got %q, %v; want %q", name, got.name, err, want)
			continue
		}
		warnings := oldFlagWarnings(fs)
		if len(warnings) != 1 || !strings.Contains(warnings[0], "'vaporwair "+want+"'") {
			t.Errorf("-%s: warnings = %q", name, warnings)
		}
	}
}

// The old flags work but aren't advertised.
func TestUsageHidesOldFlags(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stderr
	os.Stderr = w
	printUsage(newFlagSet())
	os.Stderr = old
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	if out := buf.String(); strings.Contains(out, "Deprecated") || strings.Contains(out, "  -w") {
		t.Errorf("usage lists the old flags:\n%s", out)
	}
}
