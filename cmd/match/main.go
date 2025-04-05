package main

import (
	"os"

	"github.com/alikarimi999/shahboard/matchservice"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

func main() {
	file := os.Getenv("CONFIG_FILE")
	if file == "" {
		file = "./deploy/match/development/config.json"
	}

	cfg := &matchservice.Config{}
	if err := utils.LoadConfigs(file, cfg); err != nil {
		panic(err)
	}

	app, err := matchservice.SetupApplication(*cfg)
	if err != nil {
		panic(err)
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
