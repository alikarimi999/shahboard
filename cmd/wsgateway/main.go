package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alikarimi999/shahboard/wsgateway"
)

func main() {
	config, err := loadConfig("deploy/wsgateway/development/config.json")
	if err != nil {
		panic(err)
	}

	app, err := wsgateway.SetupApplication(config)
	if err != nil {
		panic(err)
	}

	app.Run()
}

func loadConfig(file string) (wsgateway.Config, error) {
	configFile, err := os.Open(file)
	if err != nil {
		return wsgateway.Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	var config wsgateway.Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return wsgateway.Config{}, fmt.Errorf("failed to decode JSON config: %w", err)
	}

	return config, nil
}
