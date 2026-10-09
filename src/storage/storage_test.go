package storage

import (
	"encoding/json"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// sandboxHome points HOME at a temp directory and unsets the XDG variables,
// so nothing touches the real home directory. It returns the temp home.
func sandboxHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	return home
}

// noTerminal replaces stdin with a pipe for the test, so Capture never prompts.
func noTerminal(t *testing.T) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old; r.Close() })
}

// Config files are created with correct structure and can be read back.
func TestCreateAndGetConfig(t *testing.T) {
	apiKey := "test-api-key-12345"
	configPath := filepath.Join(t.TempDir(), ConfigFileName)

	if err := CreateConfig(configPath, apiKey); err != nil {
		t.Fatalf("CreateConfig failed: %v", err)
	}

	config, err := GetConfig(configPath)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if config.AirNowAPIKey != apiKey {
		t.Errorf("Config API key mismatch: got %q, want %q", config.AirNowAPIKey, apiKey)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	var jsonCheck Config
	if err := json.Unmarshal(data, &jsonCheck); err != nil {
		t.Errorf("Config file is not valid JSON: %v", err)
	}
}

// A broken config file is an error that says how to recover, not a crash.
func TestGetConfigInvalidJSON(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`{"airnowapikey": "abc`), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := GetConfig(configPath)
	if err == nil {
		t.Fatal("GetConfig accepted invalid JSON")
	}
	if !strings.Contains(err.Error(), "delete it to run setup again") {
		t.Errorf("error = %q, want a recovery hint", err)
	}
}

// Forecasts and call info round-trip through the cache, including whether the location
// came from IP lookup.
func TestSaveForecastsRoundTrip(t *testing.T) {
	a := AppConfig{CacheDir: t.TempDir()}
	coords := geolocation.Coordinates{Latitude: "40.7128", Longitude: "-74.0060", City: "New York", Zip: "10001"}
	wf := weather.Forecast{
		Latitude: 40.7128,
		Timezone: "America/New_York",
		Hourly:   weather.DataBlock{Data: []weather.DataPoint{{Time: 1730000000, Temperature: 72}, {Time: 1730003600, Temperature: 71}}},
		Daily:    weather.DataBlock{Data: []weather.DataPoint{{Time: 1730000000, TemperatureMax: 75, TemperatureMin: 60}}},
	}
	af := []air.Forecast{
		{DateForecast: "2025-11-01", ParameterName: "O3", AQI: 55, Category: air.Category{Number: 2, Name: "Moderate"}},
		{DateForecast: "2025-11-01", ParameterName: "PM2.5", AQI: 35, Category: air.Category{Number: 1, Name: "Good"}},
	}

	before := time.Now()
	if err := SaveForecasts(a, coords, true, wf, af); err != nil {
		t.Fatal(err)
	}
	after := time.Now()

	gotW, err := LoadSavedWeather(a.CacheFile(SavedWeatherFileName))
	if err != nil {
		t.Fatal(err)
	}
	if gotW.Latitude != wf.Latitude || gotW.Timezone != wf.Timezone || len(gotW.Hourly.Data) != 2 || gotW.Daily.Data[0].TemperatureMin != 60 {
		t.Errorf("weather = %+v, want %+v", gotW, wf)
	}

	gotA, err := LoadSavedAir(a.CacheFile(SavedAirFileName))
	if err != nil {
		t.Fatal(err)
	}
	if len(gotA) != 2 || gotA[1].ParameterName != "PM2.5" || gotA[1].AQI != 35 {
		t.Errorf("air = %+v, want %+v", gotA, af)
	}

	pc, err := LoadCallInfo(a.CacheFile(SavedCallFileName))
	if err != nil {
		t.Fatal(err)
	}
	if pc.Coordinates != coords || !pc.ByIP {
		t.Errorf("call info = %+v, want %+v by IP", pc, coords)
	}
	if pc.Time.Before(before) || pc.Time.After(after) {
		t.Errorf("call time %v, want between %v and %v", pc.Time, before, after)
	}
}

// The config file holds the API key and must be owner-only; forecasts need not be.
func TestFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix permission test on Windows")
	}
	dir := t.TempDir()

	configPath := filepath.Join(dir, ConfigFileName)
	if err := CreateConfig(configPath, "secret-api-key"); err != nil {
		t.Fatalf("CreateConfig failed: %v", err)
	}
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0600 {
		t.Errorf("Config file has insecure permissions: got %#o, want 0600 (owner-only)", perm)
	}

	a := AppConfig{CacheDir: dir}
	if err := SaveForecasts(a, geolocation.Coordinates{}, false, weather.Forecast{}, nil); err != nil {
		t.Fatal(err)
	}
	fileInfo, err = os.Stat(a.CacheFile(SavedWeatherFileName))
	if err != nil {
		t.Fatalf("Failed to stat forecast file: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0644 {
		t.Errorf("Forecast file has unexpected permissions: got %#o, want 0644", perm)
	}
}

// An unreadable cache is an error the caller can handle.
func TestLoadSavedWeatherError(t *testing.T) {
	path := filepath.Join(t.TempDir(), SavedWeatherFileName)
	if _, err := LoadSavedWeather(path); err == nil {
		t.Error("expected an error for a missing cache file")
	}
	if err := os.WriteFile(path, []byte("{truncated"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSavedWeather(path); err == nil {
		t.Error("expected an error for a truncated cache file")
	}
}

// Updating the default zip keeps the config owner-only.
func TestUpdateDefaultZipCodePreservesPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix permission test on Windows")
	}
	configPath := filepath.Join(t.TempDir(), ConfigFileName)
	if err := CreateConfig(configPath, "test-key"); err != nil {
		t.Fatalf("CreateConfig failed: %v", err)
	}

	if err := UpdateDefaultZipCode(configPath, "10001"); err != nil {
		t.Fatalf("UpdateDefaultZipCode failed: %v", err)
	}

	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config after update: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0600 {
		t.Errorf("UpdateDefaultZipCode changed permissions: got %#o, want 0600", perm)
	}

	config, err := GetConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if config.DefaultZipCode != "10001" || config.AirNowAPIKey != "test-key" {
		t.Errorf("config = %+v, want zip 10001 and the key kept", config)
	}

	if err := UpdateDefaultZipCode(configPath, ""); err != nil {
		t.Fatal(err)
	}
	if config, _ := GetConfig(configPath); config.DefaultZipCode != "" {
		t.Errorf("empty zip did not clear the default: %q", config.DefaultZipCode)
	}
}

// Atomic writes rename into place and leave no temp files behind.
func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.json")
	for _, content := range []string{"first", "second"} {
		if err := writeFileAtomic(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != content {
			t.Errorf("read %q, %v; want %q", got, err, content)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("directory has %d entries, want only file.json", len(entries))
	}
}

// Config and cache follow the XDG variables, ignore relative values, and fall back to $HOME.
func TestXDGDirs(t *testing.T) {
	home := sandboxHome(t)

	if dir, _ := ConfigDir(); dir != filepath.Join(home, ".config", "vaporwair") {
		t.Errorf("ConfigDir() = %q with no XDG_CONFIG_HOME", dir)
	}
	if dir, _ := CacheDir(); dir != filepath.Join(home, ".cache", "vaporwair") {
		t.Errorf("CacheDir() = %q with no XDG_CACHE_HOME", dir)
	}

	t.Setenv("XDG_CONFIG_HOME", "/xdg/config")
	t.Setenv("XDG_CACHE_HOME", "relative/cache")
	if dir, _ := ConfigDir(); dir != filepath.Join("/xdg/config", "vaporwair") {
		t.Errorf("ConfigDir() = %q, want it under XDG_CONFIG_HOME", dir)
	}
	if dir, _ := CacheDir(); dir != filepath.Join(home, ".cache", "vaporwair") {
		t.Errorf("CacheDir() = %q, want a relative XDG_CACHE_HOME ignored", dir)
	}

	if got := Tilde(filepath.Join(home, ".config", "vaporwair")); got != "~/.config/vaporwair" {
		t.Errorf("Tilde = %q", got)
	}
	if got := Tilde(home + "x/file"); got != home+"x/file" {
		t.Errorf("Tilde abbreviated a sibling directory: %q", got)
	}
}

// The first run creates the directories and a config without prompting when stdin
// isn't a terminal.
func TestInitializeAppConfigFirstRun(t *testing.T) {
	home := sandboxHome(t)
	noTerminal(t)

	appConfig, err := InitializeAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !appConfig.FirstRun {
		t.Error("FirstRun = false on the first run")
	}
	if appConfig.ConfigFile() != filepath.Join(home, ".config", "vaporwair", "config.json") {
		t.Errorf("ConfigFile() = %q", appConfig.ConfigFile())
	}
	if _, err := os.Stat(appConfig.CacheDir); err != nil {
		t.Errorf("cache directory not created: %v", err)
	}

	again, err := InitializeAppConfig()
	if err != nil || again.FirstRun {
		t.Errorf("second run: FirstRun %v, err %v", again.FirstRun, err)
	}
}

// A config in ~/.vaporwair moves to the XDG config directory, and the old directory goes.
func TestInitializeAppConfigMigratesLegacyDir(t *testing.T) {
	home := sandboxHome(t)
	noTerminal(t)
	legacy := filepath.Join(home, ".vaporwair")
	if err := os.MkdirAll(legacy, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, ConfigFileName), []byte(`{"airnowapikey":"old-key","defaultzipcode":"05401"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, SavedWeatherFileName), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	appConfig, err := InitializeAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if appConfig.FirstRun {
		t.Error("a migrated config should not count as a first run")
	}
	if appConfig.Config.AirNowAPIKey != "old-key" || appConfig.Config.DefaultZipCode != "05401" {
		t.Errorf("migrated config = %+v", appConfig.Config)
	}
	if info, err := os.Stat(appConfig.ConfigFile()); err != nil || info.Mode().Perm() != 0600 {
		t.Errorf("migrated config file: %v, %v", info, err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("legacy directory still exists: %v", err)
	}
}

// A save that fails partway leaves no call info, so the next run fetches again instead of
// pairing the new location with the old forecast.
func TestSaveForecastsWritesCallInfoLast(t *testing.T) {
	a := AppConfig{CacheDir: t.TempDir()}
	coords := geolocation.Coordinates{Latitude: "44.47", Longitude: "-73.21", Zip: "05401"}
	if err := SaveForecasts(a, coords, false, weather.Forecast{}, []air.Forecast{}); err != nil {
		t.Fatal(err)
	}
	if pc, err := LoadCallInfo(a.CacheFile(SavedCallFileName)); err != nil || pc.Coordinates != coords {
		t.Fatalf("call info after a good save: %+v, %v", pc, err)
	}

	// A directory where the air file goes makes the second write fail.
	if err := os.Mkdir(a.CacheFile(SavedAirFileName)+".blocker", 0755); err != nil {
		t.Fatal(err)
	}
	os.Remove(a.CacheFile(SavedAirFileName))
	if err := os.Rename(a.CacheFile(SavedAirFileName)+".blocker", a.CacheFile(SavedAirFileName)); err != nil {
		t.Fatal(err)
	}
	if err := SaveForecasts(a, coords, false, weather.Forecast{}, []air.Forecast{}); err == nil {
		t.Fatal("expected the air write to fail")
	}
	if _, err := LoadCallInfo(a.CacheFile(SavedCallFileName)); err == nil {
		t.Error("call info survived a failed save")
	}
}
