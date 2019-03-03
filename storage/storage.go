package storage

import (
	"encoding/json"
	"fmt"
	"github.com/jeff-bruemmer/vaporwair/air"
	"github.com/jeff-bruemmer/vaporwair/geolocation"
	"github.com/jeff-bruemmer/vaporwair/weather"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"time"
)

const VaporwairDir = "/.vaporwair/"
const SavedWeatherFileName = VaporwairDir + "weather-forecast.json"
const SavedAirFileName = VaporwairDir + "air-forecast.json"
const ConfigFileName = VaporwairDir + "config.json"
const SavedCallFileName = VaporwairDir + "last-call.json"

// The Config type is used to store API keys.
type Config struct {
	DarkSkyAPIKey string `json:"darkskyapikey"`
	AirNowAPIKey  string `json:"airnowapikey"`
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

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
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
	d, err := exists(path)
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
	b, err := ioutil.ReadFile(path)
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
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading forecast from disk.", err)
	}
	err = json.Unmarshal(b, &f)
	if err != nil {
		log.Fatal("Error unmarshalling json into Forecast.", err)
	}
	return f, nil
}

// Checks home folder for vaporwair config file to retrieve API keys.
func GetConfig(filepath string) Config {
	configFile, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Could not find config file in home directory.")
		log.Fatal(err)
	}
	defer configFile.Close()
	var config Config
	bytes, _ := ioutil.ReadAll(configFile)
	// Validate json data
	valid := json.Valid(bytes)
	if !valid {
		log.Fatal("\nThe config file:\n", filepath, "\ndoes not contain valid JSON.")
	}
	json.Unmarshal(bytes, &config)
	return config
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
	f, err := ioutil.ReadFile(path)
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
	err = ioutil.WriteFile(path, c, 0644)
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
	err = ioutil.WriteFile(path, c, 0644)
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
	err = ioutil.WriteFile(path, c, 0644)
	if err != nil {
		return false
	}
	return true
}
