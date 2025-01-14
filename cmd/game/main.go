package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alikarimi999/shahboard/gameservice"
)

func main() {
	config, err := loadConfig("deploy/game/development/config.json")
	if err != nil {
		panic(err)
	}
	app, err := gameservice.SetupApplication(config)
	if err != nil {
		panic(err)
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
func loadConfig(file string) (gameservice.Config, error) {
	configFile, err := os.Open(file)
	if err != nil {
		return gameservice.Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	var config gameservice.Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return gameservice.Config{}, fmt.Errorf("failed to decode JSON config: %w", err)
	}

	return config, nil
}
