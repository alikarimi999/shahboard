package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alikarimi999/shahboard/matchservice"
)

func main() {
	config, err := loadConfig("deploy/match/development/config.json")
	if err != nil {
		panic(err)
	}

	app, err := matchservice.SetupApplication(config)
	if err != nil {
		panic(err)
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func loadConfig(file string) (matchservice.Config, error) {
	configFile, err := os.Open(file)
	if err != nil {
		return matchservice.Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	var config matchservice.Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return matchservice.Config{}, fmt.Errorf("failed to decode JSON config: %w", err)
	}

	return config, nil
}
