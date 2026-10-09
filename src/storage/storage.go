// This package contains OS utilites for storing and retrieving payloads from API calls.
package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Application configuration and caching constants.
// Config lives in the XDG config directory, cached forecasts in the XDG cache directory.
const (
	AppDirName           = "vaporwair"
	ConfigFileName       = "config.json"
	SavedWeatherFileName = "weather-forecast.json"
	SavedAirFileName     = "air-forecast.json"
	SavedCallFileName    = "last-call.json"
	CacheTimeoutMinutes  = 5

	// legacyDirName is ~/.vaporwair, which held config and cache before vaporwair
	// followed the XDG base directories. InitializeAppConfig moves its config once.
	legacyDirName = ".vaporwair"
)

// Config stores API keys and settings (NOAA doesn't require an API key).
type Config struct {
	AirNowAPIKey   string `json:"airnowapikey"`
	DefaultZipCode string `json:"defaultzipcode,omitempty"`
}

// AppConfig holds runtime configuration for the application.
type AppConfig struct {
	Config              Config
	ConfigDir           string
	CacheDir            string
	CacheTimeoutMinutes float64
	// FirstRun is true when this run created the config file.
	FirstRun bool
}

// ConfigFile is the path of config.json.
func (a AppConfig) ConfigFile() string { return filepath.Join(a.ConfigDir, ConfigFileName) }

// CacheFile is the path of a cached file such as SavedWeatherFileName.
func (a AppConfig) CacheFile(name string) string { return filepath.Join(a.CacheDir, name) }

// APICallInfo contains metadata to determine validity of last API call.
// It is written after the forecasts it describes, so finding it means they are complete.
type APICallInfo struct {
	Time        time.Time
	Coordinates geolocation.Coordinates
	// ByIP is true when the coordinates came from IP lookup rather than a zip code.
	ByIP bool `json:",omitempty"`
}

// ConfigDir returns $XDG_CONFIG_HOME/vaporwair, or ~/.config/vaporwair.
func ConfigDir() (string, error) { return xdgDir("XDG_CONFIG_HOME", ".config") }

// CacheDir returns $XDG_CACHE_HOME/vaporwair, or ~/.cache/vaporwair.
func CacheDir() (string, error) { return xdgDir("XDG_CACHE_HOME", ".cache") }

// xdgDir follows the XDG base directory spec on every platform, so the config file is in
// the same easy-to-type place on macOS as on Linux. The spec says to ignore relative paths.
func xdgDir(env, fallback string) (string, error) {
	if dir := os.Getenv(env); filepath.IsAbs(dir) {
		return filepath.Join(dir, AppDirName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	return filepath.Join(home, fallback, AppDirName), nil
}

// Tilde abbreviates the home directory at the start of path as "~", for messages.
func Tilde(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if rest, ok := strings.CutPrefix(path, home); ok && (rest == "" || rest[0] == filepath.Separator) {
		return "~" + rest
	}
	return path
}

// Capture takes a prompt and returns user-entered string. The prompt goes to stderr so it
// never lands in a piped report, and with no terminal on stdin it returns "" without asking.
func Capture(prompt string) string {
	if info, err := os.Stdin.Stat(); err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return ""
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprint(os.Stderr, prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// CreateConfig writes a new config file at path holding the AirNow API key.
// The file is 0600 because it holds the key.
func CreateConfig(path, anak string) error {
	return saveJSON(path, Config{AirNowAPIKey: anak}, 0600)
}

// loadJSON reads a cached file. A missing or unreadable cache is an error for the caller
// to handle, not a fatal one.
func loadJSON[T any](path string) (T, error) {
	var v T
	b, err := os.ReadFile(path)
	if err != nil {
		return v, err
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return v, fmt.Errorf("reading cached %s: %w", path, err)
	}
	return v, nil
}

// LoadSavedWeather loads the cached weather forecast.
func LoadSavedWeather(path string) (weather.Forecast, error) { return loadJSON[weather.Forecast](path) }

// LoadSavedAir loads the cached air quality forecast.
func LoadSavedAir(path string) ([]air.Forecast, error) { return loadJSON[[]air.Forecast](path) }

// LoadCallInfo loads the call info that says when and where the cache was fetched.
func LoadCallInfo(path string) (APICallInfo, error) { return loadJSON[APICallInfo](path) }

// GetConfig loads API keys and settings from the config file.
func GetConfig(path string) (Config, error) {
	var config Config
	b, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("reading config file: %w", err)
	}
	if err := json.Unmarshal(b, &config); err != nil {
		return config, fmt.Errorf("config file %s is not valid JSON (%v); fix it, or delete it to run setup again", Tilde(path), err)
	}
	config.AirNowAPIKey = strings.TrimSpace(config.AirNowAPIKey)
	return config, nil
}

// InitializeAppConfig sets up and returns the complete application configuration.
// It creates the config and cache directories and, on first run, the config file.
func InitializeAppConfig() (AppConfig, error) {
	appConfig := AppConfig{CacheTimeoutMinutes: CacheTimeoutMinutes}

	var err error
	if appConfig.ConfigDir, err = ConfigDir(); err != nil {
		return appConfig, err
	}
	if appConfig.CacheDir, err = CacheDir(); err != nil {
		return appConfig, err
	}
	// The config directory is private: the config file holds the API key.
	if err := os.MkdirAll(appConfig.ConfigDir, 0700); err != nil {
		return appConfig, fmt.Errorf("creating config directory: %w", err)
	}
	if err := os.MkdirAll(appConfig.CacheDir, 0755); err != nil {
		return appConfig, fmt.Errorf("creating cache directory: %w", err)
	}

	configFile := appConfig.ConfigFile()
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		moved, err := migrateLegacyConfig(configFile)
		if err != nil {
			return appConfig, err
		}
		if !moved {
			appConfig.FirstRun = true
			key := Capture("Enter AirNow API key (press Enter to skip): ")
			if err := CreateConfig(configFile, key); err != nil {
				return appConfig, fmt.Errorf("creating config file: %w", err)
			}
		}
	}

	appConfig.Config, err = GetConfig(configFile)
	return appConfig, err
}

// migrateLegacyConfig moves config.json out of ~/.vaporwair and deletes the old cache
// files, then the directory if nothing else is in it. It reports whether a config moved.
func migrateLegacyConfig(configFile string) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, nil
	}
	legacy := filepath.Join(home, legacyDirName)
	data, err := os.ReadFile(filepath.Join(legacy, ConfigFileName))
	if err != nil {
		return false, nil // nothing to move
	}
	if err := writeFileAtomic(configFile, data, 0600); err != nil {
		return false, fmt.Errorf("moving config from %s: %w", Tilde(legacy), err)
	}
	for _, name := range []string{ConfigFileName, SavedWeatherFileName, SavedAirFileName, SavedCallFileName} {
		os.Remove(filepath.Join(legacy, name))
	}
	os.Remove(legacy) // fails, harmlessly, if the user kept other files there
	fmt.Fprintf(os.Stderr, "Note: moved config from %s to %s\n", Tilde(legacy), Tilde(configFile))
	return true, nil
}

// UpdateDefaultZipCode saves the zip code as the default in the config file at path.
// An empty zip code clears the default.
func UpdateDefaultZipCode(path string, zipCode string) error {
	config, err := GetConfig(path)
	if err != nil {
		return err
	}
	config.DefaultZipCode = zipCode
	return saveJSON(path, config, 0600)
}

// SaveForecasts caches both forecasts and the call info that vouches for them. The call
// info is removed first and written last, so a save that fails or is interrupted midway
// leaves no cache, rather than one that pairs a location with another place's forecast.
func SaveForecasts(a AppConfig, c geolocation.Coordinates, byIP bool, wf weather.Forecast, af []air.Forecast) error {
	callFile := a.CacheFile(SavedCallFileName)
	if err := os.Remove(callFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := saveJSON(a.CacheFile(SavedWeatherFileName), wf, 0644); err != nil {
		return err
	}
	if err := saveJSON(a.CacheFile(SavedAirFileName), af, 0644); err != nil {
		return err
	}
	return SaveCall(callFile, APICallInfo{Time: time.Now(), Coordinates: c, ByIP: byIP})
}

// SaveCall saves info for future calls.
// Call info is public data (coordinates, timestamp), so 0644 is acceptable.
func SaveCall(path string, info APICallInfo) error {
	return saveJSON(path, info, 0644)
}

// saveJSON marshals data to JSON and saves it to a file.
func saveJSON(path string, data any, perm os.FileMode) error {
	content, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}
	return writeFileAtomic(path, content, perm)
}

// writeFileAtomic replaces path with data in one step: it writes a temp file beside path
// and renames it into place, so an interrupted write never leaves a truncated file and
// another vaporwair reading at the same time sees the old file or the new one.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op once renamed
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
