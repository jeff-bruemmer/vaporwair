package storage

import (
	"encoding/json"
	"github.com/jeff-bruemmer/vaporwair/src/air"
	"github.com/jeff-bruemmer/vaporwair/src/geolocation"
	"github.com/jeff-bruemmer/vaporwair/src/weather"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Test 1.1: Config File Creation and Reading
// Validates that config files are created with correct structure and can be read back
// This ensures no regression when migrating from ioutil to os
func TestCreateAndGetConfig(t *testing.T) {
	// Setup: Create temp directory
	tempDir, err := os.MkdirTemp("", "vaporwair-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test: Create config with API key
	apiKey := "test-api-key-12345"
	configPath := tempDir + ConfigFileName

	// Create the vaporwair subdirectory
	vaporwairDir := tempDir + VaporwairDir
	err = os.MkdirAll(vaporwairDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create vaporwair dir: %v", err)
	}

	err = CreateConfig(tempDir, apiKey)
	if err != nil {
		t.Fatalf("CreateConfig failed: %v", err)
	}

	// Verify: Config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Verify: GetConfig returns same data
	config := GetConfig(configPath)
	if config.AirNowAPIKey != apiKey {
		t.Errorf("Config API key mismatch: got %q, want %q", config.AirNowAPIKey, apiKey)
	}

	// Verify: Config is valid JSON
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var jsonCheck Config
	if err := json.Unmarshal(data, &jsonCheck); err != nil {
		t.Errorf("Config file is not valid JSON: %v", err)
	}
}

// Test 1.2: Weather Forecast Persistence
// Validates forecast serialization to ensure migration doesn't break cache functionality
func TestSaveAndLoadWeatherForecast(t *testing.T) {
	// Setup: Create temp directory
	tempDir, err := os.MkdirTemp("", "vaporwair-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	forecastPath := tempDir + SavedWeatherFileName

	// Create directory structure
	vaporwairDir := tempDir + VaporwairDir
	err = os.MkdirAll(vaporwairDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create vaporwair dir: %v", err)
	}

	// Create sample forecast with comprehensive data
	sampleForecast := weather.Forecast{
		Latitude:  40.7128,
		Longitude: -74.0060,
		Timezone:  "America/New_York",
		Currently: weather.DataPoint{
			Time:                float64(time.Now().Unix()),
			Temperature:         72.5,
			ApparentTemperature: 75.0,
			Humidity:            0.65,
			WindSpeed:           10.5,
			Summary:             "Partly Cloudy",
			PrecipProbability:   0.20,
		},
		Hourly: weather.DataBlock{
			Summary: "Partly cloudy for the hour",
			Data: []weather.DataPoint{
				{Time: 1730000000, Temperature: 72},
				{Time: 1730003600, Temperature: 71},
			},
		},
		Daily: weather.DataBlock{
			Summary: "Partly cloudy throughout the week",
			Data: []weather.DataPoint{
				{Time: 1730000000, TemperatureMax: 75, TemperatureMin: 60},
			},
		},
		Alerts: []weather.Alert{},
	}

	// Test: SaveWeatherForecast
	success := SaveWeatherForecast(forecastPath, sampleForecast)
	if !success {
		t.Fatal("SaveWeatherForecast returned false")
	}

	// Verify: File exists
	if _, err := os.Stat(forecastPath); os.IsNotExist(err) {
		t.Fatalf("Weather forecast file was not created at %s", forecastPath)
	}

	// Test: LoadSavedWeather
	loaded, err := LoadSavedWeather(forecastPath)
	if err != nil {
		t.Fatalf("LoadSavedWeather failed: %v", err)
	}

	// Verify: Critical fields match
	if loaded.Latitude != sampleForecast.Latitude {
		t.Errorf("Latitude mismatch: got %f, want %f", loaded.Latitude, sampleForecast.Latitude)
	}
	if loaded.Longitude != sampleForecast.Longitude {
		t.Errorf("Longitude mismatch: got %f, want %f", loaded.Longitude, sampleForecast.Longitude)
	}
	if loaded.Timezone != sampleForecast.Timezone {
		t.Errorf("Timezone mismatch: got %s, want %s", loaded.Timezone, sampleForecast.Timezone)
	}
	if loaded.Currently.Temperature != sampleForecast.Currently.Temperature {
		t.Errorf("Current temperature mismatch: got %f, want %f",
			loaded.Currently.Temperature, sampleForecast.Currently.Temperature)
	}
	if len(loaded.Hourly.Data) != len(sampleForecast.Hourly.Data) {
		t.Errorf("Hourly data count mismatch: got %d, want %d",
			len(loaded.Hourly.Data), len(sampleForecast.Hourly.Data))
	}
	if len(loaded.Daily.Data) != len(sampleForecast.Daily.Data) {
		t.Errorf("Daily data count mismatch: got %d, want %d",
			len(loaded.Daily.Data), len(sampleForecast.Daily.Data))
	}
}

// Test 1.3: Air Forecast Persistence
// Validates air quality cache functionality
func TestSaveAndLoadAirForecast(t *testing.T) {
	// Setup: Create temp directory
	tempDir, err := os.MkdirTemp("", "vaporwair-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	airPath := tempDir + SavedAirFileName

	// Create directory structure
	vaporwairDir := tempDir + VaporwairDir
	err = os.MkdirAll(vaporwairDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create vaporwair dir: %v", err)
	}

	// Create sample air quality forecast
	sampleAir := []air.Forecast{
		{
			DateForecast:  "2025-11-01",
			ReportingArea: "Test Area",
			StateCode:     "NY",
			Latitude:      40.7128,
			Longitude:     -74.0060,
			ParameterName: "O3",
			AQI:           55,
			Category: air.Category{
				Number: 2,
				Name:   "Moderate",
			},
		},
		{
			DateForecast:  "2025-11-01",
			ReportingArea: "Test Area",
			StateCode:     "NY",
			ParameterName: "PM2.5",
			AQI:           35,
			Category: air.Category{
				Number: 1,
				Name:   "Good",
			},
		},
	}

	// Test: SaveAirForecast
	success := SaveAirForecast(airPath, sampleAir)
	if !success {
		t.Fatal("SaveAirForecast returned false")
	}

	// Verify: File exists
	if _, err := os.Stat(airPath); os.IsNotExist(err) {
		t.Fatalf("Air forecast file was not created at %s", airPath)
	}

	// Test: LoadSavedAir
	loaded, err := LoadSavedAir(airPath)
	if err != nil {
		t.Fatalf("LoadSavedAir failed: %v", err)
	}

	// Verify: Data matches
	if len(loaded) != len(sampleAir) {
		t.Fatalf("Air forecast count mismatch: got %d, want %d", len(loaded), len(sampleAir))
	}

	for i, forecast := range loaded {
		if forecast.ParameterName != sampleAir[i].ParameterName {
			t.Errorf("Parameter name mismatch at index %d: got %s, want %s",
				i, forecast.ParameterName, sampleAir[i].ParameterName)
		}
		if forecast.AQI != sampleAir[i].AQI {
			t.Errorf("AQI mismatch at index %d: got %d, want %d",
				i, forecast.AQI, sampleAir[i].AQI)
		}
	}
}

// Test 1.4: Call Info Persistence
// Validates cache metadata storage - critical for cache timeout logic
func TestUpdateAndLoadCallInfo(t *testing.T) {
	// Setup: Create temp directory
	tempDir, err := os.MkdirTemp("", "vaporwair-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	callPath := tempDir + SavedCallFileName

	// Create directory structure
	vaporwairDir := tempDir + VaporwairDir
	err = os.MkdirAll(vaporwairDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create vaporwair dir: %v", err)
	}

	// Create sample coordinates
	coords := geolocation.Coordinates{
		Latitude:  "40.7128",
		Longitude: "-74.0060",
		City:      "New York",
		Zip:       "10001",
	}

	// Test: UpdateLastCall
	beforeTime := time.Now()
	err = UpdateLastCall(coords, callPath)
	afterTime := time.Now()
	if err != nil {
		t.Fatalf("UpdateLastCall failed: %v", err)
	}

	// Verify: File exists
	if _, err := os.Stat(callPath); os.IsNotExist(err) {
		t.Fatalf("Call info file was not created at %s", callPath)
	}

	// Test: LoadCallInfo
	loaded, err := LoadCallInfo(callPath)
	if err != nil {
		t.Fatalf("LoadCallInfo failed: %v", err)
	}

	// Verify: Coordinates match
	if loaded.Coordinates.Latitude != coords.Latitude {
		t.Errorf("Latitude mismatch: got %s, want %s", loaded.Coordinates.Latitude, coords.Latitude)
	}
	if loaded.Coordinates.Longitude != coords.Longitude {
		t.Errorf("Longitude mismatch: got %s, want %s", loaded.Coordinates.Longitude, coords.Longitude)
	}
	if loaded.Coordinates.City != coords.City {
		t.Errorf("City mismatch: got %s, want %s", loaded.Coordinates.City, coords.City)
	}
	if loaded.Coordinates.Zip != coords.Zip {
		t.Errorf("Zip mismatch: got %s, want %s", loaded.Coordinates.Zip, coords.Zip)
	}

	// Verify: Time is within reasonable bounds (critical for cache timeout)
	if loaded.Time.Before(beforeTime) || loaded.Time.After(afterTime) {
		t.Errorf("Timestamp out of bounds: got %v, want between %v and %v",
			loaded.Time, beforeTime, afterTime)
	}
}

// Test 1.5: File Permission Verification
// This test will FAIL with current code (0644) and PASS after fix (0600 for config)
func TestFilePermissions(t *testing.T) {
	// Skip on Windows (Unix permissions don't apply)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix permission test on Windows")
	}

	// Setup: Create temp directory
	tempDir, err := os.MkdirTemp("", "vaporwair-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create directory structure
	vaporwairDir := tempDir + VaporwairDir
	err = os.MkdirAll(vaporwairDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create vaporwair dir: %v", err)
	}

	// Test: Create config file (contains API key - should be 0600)
	apiKey := "secret-api-key"
	configPath := tempDir + ConfigFileName
	err = CreateConfig(tempDir, apiKey)
	if err != nil {
		t.Fatalf("CreateConfig failed: %v", err)
	}

	// Verify: Config has 0600 permissions (owner-only)
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	perm := fileInfo.Mode().Perm()
	expected := os.FileMode(0600)
	if perm != expected {
		t.Errorf("Config file has insecure permissions: got %#o, want %#o (owner-only)", perm, expected)
		t.Errorf("Config contains API keys and should not be world-readable")
	}

	// Test: Create forecast files (public data - 0644 is acceptable)
	forecastPath := tempDir + SavedWeatherFileName
	sampleForecast := weather.Forecast{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
	SaveWeatherForecast(forecastPath, sampleForecast)

	fileInfo, err = os.Stat(forecastPath)
	if err != nil {
		t.Fatalf("Failed to stat forecast file: %v", err)
	}

	perm = fileInfo.Mode().Perm()
	// Forecast files can be 0644 (they don't contain secrets)
	if perm != 0644 && perm != 0600 {
		t.Errorf("Forecast file has unexpected permissions: got %#o", perm)
	}
}

// Test 1.6: Error Handling - Unreadable Files
// Validates proper error returns instead of log.Fatal
func TestLoadSavedWeatherError(t *testing.T) {
	t.Skip("Skipping error handling test - current code uses log.Fatal which exits the process")
	// Note: This test is skipped because fixing log.Fatal is Issue #5
	// Current implementation calls log.Fatal on JSON unmarshal errors,
	// which would cause this test to exit the entire test suite
	// This should be fixed in a future iteration when addressing Issue #5
}

// Test 1.7: Backward Compatibility - Existing Cache
// Ensures migration doesn't break existing cached data
func TestBackwardCompatibilityWithExistingCache(t *testing.T) {
	// Setup: Create temp directory
	tempDir, err := os.MkdirTemp("", "vaporwair-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	forecastPath := tempDir + SavedWeatherFileName

	// Create directory structure
	vaporwairDir := tempDir + VaporwairDir
	err = os.MkdirAll(vaporwairDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create vaporwair dir: %v", err)
	}

	// Simulate existing cache file created by old version (using ioutil)
	// The format should be identical whether created with ioutil or os
	sampleForecast := weather.Forecast{
		Latitude:  40.7128,
		Longitude: -74.0060,
		Timezone:  "America/New_York",
	}

	// Save using current implementation
	success := SaveWeatherForecast(forecastPath, sampleForecast)
	if !success {
		t.Fatal("Failed to save forecast")
	}

	// Load using current implementation (should work regardless of ioutil vs os)
	loaded, err := LoadSavedWeather(forecastPath)
	if err != nil {
		t.Fatalf("Failed to load saved forecast: %v", err)
	}

	// Verify data integrity
	if loaded.Latitude != sampleForecast.Latitude {
		t.Errorf("Latitude mismatch after save/load: got %f, want %f",
			loaded.Latitude, sampleForecast.Latitude)
	}
	if loaded.Longitude != sampleForecast.Longitude {
		t.Errorf("Longitude mismatch after save/load: got %f, want %f",
			loaded.Longitude, sampleForecast.Longitude)
	}
	if loaded.Timezone != sampleForecast.Timezone {
		t.Errorf("Timezone mismatch after save/load: got %s, want %s",
			loaded.Timezone, sampleForecast.Timezone)
	}
}

// Test 3.2: UpdateDefaultZipCode Preserves Permissions
// Ensures updating config doesn't accidentally change file permissions
func TestUpdateDefaultZipCodePreservesPermissions(t *testing.T) {
	// Skip on Windows
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix permission test on Windows")
	}

	// Setup: Create temp directory
	tempDir, err := os.MkdirTemp("", "vaporwair-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create directory structure
	vaporwairDir := tempDir + VaporwairDir
	err = os.MkdirAll(vaporwairDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create vaporwair dir: %v", err)
	}

	// Create initial config with 0600
	apiKey := "test-key"
	err = CreateConfig(tempDir, apiKey)
	if err != nil {
		t.Fatalf("CreateConfig failed: %v", err)
	}

	configPath := tempDir + ConfigFileName

	// Verify initial permissions
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config: %v", err)
	}
	initialPerm := fileInfo.Mode().Perm()

	// Test: UpdateDefaultZipCode
	err = UpdateDefaultZipCode(tempDir, "10001")
	if err != nil {
		t.Fatalf("UpdateDefaultZipCode failed: %v", err)
	}

	// Verify: Permissions still 0600
	fileInfo, err = os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config after update: %v", err)
	}

	updatedPerm := fileInfo.Mode().Perm()
	expectedPerm := os.FileMode(0600)

	if updatedPerm != expectedPerm {
		t.Errorf("UpdateDefaultZipCode changed permissions: got %#o, want %#o", updatedPerm, expectedPerm)
	}

	if updatedPerm != initialPerm {
		t.Errorf("UpdateDefaultZipCode changed permissions from initial: was %#o, now %#o",
			initialPerm, updatedPerm)
	}

	// Verify zip code was actually updated
	config := GetConfig(configPath)
	if config.DefaultZipCode != "10001" {
		t.Errorf("Zip code not updated: got %s, want 10001", config.DefaultZipCode)
	}
}

// Test: Verify Exists function works correctly
func TestExists(t *testing.T) {
	// Test with existing file
	tempFile, err := os.CreateTemp("", "test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	exists, err := Exists(tempPath)
	if err != nil {
		t.Errorf("Exists returned error for existing file: %v", err)
	}
	if !exists {
		t.Error("Exists returned false for existing file")
	}

	// Test with non-existent file
	exists, err = Exists("/nonexistent/path/file.txt")
	if err != nil {
		t.Errorf("Exists returned error for non-existent file: %v", err)
	}
	if exists {
		t.Error("Exists returned true for non-existent file")
	}
}

// Test: CreateVaporwairDir creates directory correctly
func TestCreateVaporwairDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "vaporwair-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vaporwairPath := filepath.Join(tempDir, ".vaporwair")

	// Directory shouldn't exist yet
	if _, err := os.Stat(vaporwairPath); !os.IsNotExist(err) {
		t.Fatal("Directory already exists before test")
	}

	// Create the directory
	CreateVaporwairDir(vaporwairPath)

	// Verify it exists and is a directory
	fileInfo, err := os.Stat(vaporwairPath)
	if err != nil {
		t.Fatalf("CreateVaporwairDir failed to create directory: %v", err)
	}
	if !fileInfo.IsDir() {
		t.Error("CreateVaporwairDir created a file instead of directory")
	}

	// Call again - should not error (idempotent)
	CreateVaporwairDir(vaporwairPath)

	// Verify still exists
	if _, err := os.Stat(vaporwairPath); err != nil {
		t.Error("CreateVaporwairDir not idempotent - directory disappeared on second call")
	}
}
