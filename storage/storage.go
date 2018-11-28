package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

type Config struct {
	DarkSkyAPIKey string `json:"darkskyapikey"`
	AirNowAPIKey  string `json:"airnowapikey"`
}

func GetHomeDir() (string, error) {
	usr, err := user.Current()
	return usr.HomeDir, err
}

func ConfigFilePath(homeDir string) string {
	return homeDir + "/.vaporwair-config.json"
}

// Checks home folder for vaporwair config file to retrieve API Keys
// Takes a home directory and a ConfigFilePath function
func GetConfig(homeDir string, cfp func(string) string) Config {
	config, err := os.Open(cfp(homeDir))
	if err != nil {
		// TODO create custom error to print instructions for creating config file
		fmt.Println("Could not find config file in home directory.")
		log.Fatal(err)
	}
	defer config.Close()
	bytes, _ := ioutil.ReadAll(config)
	var c Config
	json.Unmarshal(bytes, &c)
	return c
}
