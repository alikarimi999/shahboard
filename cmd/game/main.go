package main

import (
	"os"

	"github.com/alikarimi999/shahboard/gameservice"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

func main() {
	file := os.Getenv("CONFIG_FILE")
	if file == "" {
		file = "./deploy/game/development/config.json"
	}

	cfg := &gameservice.Config{}
	if err := utils.LoadConfigs(file, cfg); err != nil {
		panic(err)
	}

	app, err := gameservice.SetupApplication(*cfg)
	if err != nil {
		panic(err)
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
