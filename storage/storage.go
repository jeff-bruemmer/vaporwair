package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

const SavedForecastFileName = "/.vaporwair-saved-forecast.json"
const ConfigFileName = "/.vaporwair-config.json"

type Config struct {
	DarkSkyAPIKey string `json:"darkskyapikey"`
	AirNowAPIKey  string `json:"airnowapikey"`
}

func GetHomeDir() (string, error) {
	usr, err := user.Current()
	return usr.HomeDir, err
}

func ConfigFilePath(homeDir string, configFileName string) string {
	return homeDir + configFileName
}

func GetSavedForecast(f string) (string, error) {
	forecast, err := os.Open(f)
	if err != nil {
		fmt.Println("No saved forecast found.")
		return "", err
	}
	forecast.Close()
	return "Saved Forecast data", nil
}

// Checks home folder for vaporwair config file to retrieve API Keys
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
