// This package contains OS utilites for storing and retrieving payloads from API calls.
package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
	"time"
)

// Application configuration constants for caching and storage.
//
// Caching Strategy:
// Vaporwair uses optimistic caching to reduce API calls and improve response times.
// Forecasts are cached to disk with metadata (timestamp, coordinates) and served
// from cache when:
//   - Less than CacheTimeoutMinutes have elapsed since last fetch
//   - User location hasn't changed
//
// This approach assumes weather forecasts don't change frequently enough to
// warrant fetching fresh data on every request within the timeout window.
const (
	VaporwairDir           = "/.vaporwair/"
	SavedWeatherFileName   = VaporwairDir + "weather-forecast.json"
	SavedAirFileName       = VaporwairDir + "air-forecast.json"
	ConfigFileName         = VaporwairDir + "config.json"
	SavedCallFileName      = VaporwairDir + "last-call.json"
	CacheTimeoutMinutes    = 5  // How long cached forecasts remain valid
)

// Config stores API keys and application settings.
// Note: NOAA API does not require an API key.
type Config struct {
	AirNowAPIKey   string `json:"airnowapikey"`
	DefaultZipCode string `json:"defaultzipcode,omitempty"`
}

// AppConfig holds runtime configuration for the application.
type AppConfig struct {
	Config             Config
	HomeDir            string
	CacheTimeoutMinutes float64
}

// APICallInfo contains metadata to determine validity of last API call.
type APICallInfo struct {
	Time        time.Time
	Coordinates geolocation.Coordinates
}

// Determines home directory in order to create vaporwair
// directory to cache forecasts and call data.
func GetHomeDir() (string, error) {
	usr, err := user.Current()
	return usr.HomeDir, err
}

// Capture takes a prompt and returns user-entered string.
func Capture(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func CreateConfig(homeDir, anak string) error {
	path := homeDir + ConfigFileName
	config := Config{}
	config.AirNowAPIKey = anak
	c, err := json.Marshal(config)
	if err != nil {
		fmt.Println("There was an error marshalling the configuration.")
		return err
	}
	// Use 0600 permissions for config file (contains API keys)
	// os.WriteFile will create the file if it doesn't exist
	err = os.WriteFile(path, c, 0600)
	if err != nil {
		fmt.Println("There was an error writing the config file to ", path)
		return err
	}
	return nil
}

// exists returns whether the given file or directory exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// Creates a directory to cache forecasts and call data
// if that directory does not already exist.
func CreateVaporwairDir(path string) {
	d, err := Exists(path)
	if err != nil {
		fmt.Println("There was a problem identifying Vaporwair directory.")
	}
	if d {
		return
	} else {
		os.Mkdir(path, 0755)
	}
}

// Loads previous weather forecast.
func LoadSavedWeather(path string) (weather.Forecast, error) {
	var f weather.Forecast
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading forecast from disk.", err)
	}
	err = json.Unmarshal(b, &f)
	if err != nil {
		log.Fatal("Error unmarshalling json into Forecast.", err)
	}
	return f, nil
}

// Loads previous air quality forecast.
func LoadSavedAir(path string) ([]air.Forecast, error) {
	var f []air.Forecast
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading forecast from disk.", err)
	}
	err = json.Unmarshal(b, &f)
	if err != nil {
		log.Fatal("Error unmarshalling json into Forecast.", err)
	}
	return f, nil
}

// GetConfig loads API keys from the config file.
func GetConfig(filepath string) Config {
	configFile, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Could not find config file in home directory.")
		log.Fatal(err)
	}
	defer configFile.Close()
	var config Config
	bytes, _ := io.ReadAll(configFile)
	// Validate json data
	valid := json.Valid(bytes)
	if !valid {
		log.Fatal("\nThe config file:\n", filepath, "\ndoes not contain valid JSON.")
	}
	json.Unmarshal(bytes, &config)

	// Trim whitespace from API key
	config.AirNowAPIKey = strings.TrimSpace(config.AirNowAPIKey)

	return config
}

// InitializeAppConfig sets up and returns the complete application configuration.
// It creates necessary directories and files if they don't exist.
func InitializeAppConfig() (AppConfig, error) {
	var appConfig AppConfig

	homeDir, err := GetHomeDir()
	if err != nil {
		return appConfig, fmt.Errorf("unable to determine home directory: %w", err)
	}

	appConfig.HomeDir = homeDir
	appConfig.CacheTimeoutMinutes = CacheTimeoutMinutes

	// Create vaporwair directory if it doesn't exist
	CreateVaporwairDir(homeDir + VaporwairDir)

	// Check if configuration file exists
	configFile := homeDir + ConfigFileName
	configExists, _ := Exists(configFile)

	// If config doesn't exist, create it with user input
	if !configExists {
		ANAPIKey := Capture("Enter Air Now API key: ")
		err := CreateConfig(homeDir, ANAPIKey)
		if err != nil {
			return appConfig, fmt.Errorf("error creating configuration: %w", err)
		}
	}

	// Load API keys
	appConfig.Config = GetConfig(configFile)

	return appConfig, nil
}

// UpdateDefaultZipCode saves the zip code as the default in the config file.
func UpdateDefaultZipCode(homeDir string, zipCode string) error {
	configFile := homeDir + ConfigFileName
	config := GetConfig(configFile)

	// Update the default zip code
	config.DefaultZipCode = zipCode

	// Save back to file
	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	// Use 0600 permissions for config file (contains API keys)
	err = os.WriteFile(configFile, configData, 0600)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

func UpdateLastCall(c geolocation.Coordinates, path string) error {
	// After call, save report.
	newCallInfo := APICallInfo{
		Time:        time.Now(),
		Coordinates: c,
	}
	err := SaveCall(path, newCallInfo)
	if err != nil {
		fmt.Println("Error saving call info.\n", err)
		return err
	} else {
		return nil
	}
}

// Loads call information to determine whether
// to retrieve forecast from server or disk
func LoadCallInfo(path string) (APICallInfo, error) {
	var lastCall APICallInfo
	f, err := os.ReadFile(path)
	if err != nil {
		return lastCall, err
	}
	err = json.Unmarshal(f, &lastCall)
	if err != nil {
		fmt.Println("Error unmarshalling last api call.\n", err)
		return lastCall, err
	}
	return lastCall, nil
}

// Save info for future calls
func SaveCall(path string, info APICallInfo) error {
	c, err := json.Marshal(info)
	if err != nil {
		return err
	}
	// Call info is public data (coordinates, timestamp), 0644 is acceptable
	err = os.WriteFile(path, c, 0644)
	if err != nil {
		return err
	}
	return nil
}

func SaveWeatherForecast(path string, f weather.Forecast) bool {
	c, err := json.Marshal(f)
	if err != nil {
		fmt.Println("Error marshalling weather forecast before saving.\n", err)
		return false
	}
	// Weather forecast is public data, 0644 is acceptable
	err = os.WriteFile(path, c, 0644)
	if err != nil {
		return false
	}
	return true
}

func SaveAirForecast(path string, a []air.Forecast) bool {
	c, err := json.Marshal(a)
	if err != nil {
		fmt.Println("Error marshalling air forecast before saving.\n", err)
		return false
	}
	// Air quality forecast is public data, 0644 is acceptable
	err = os.WriteFile(path, c, 0644)
	if err != nil {
		return false
	}
	return true
}
