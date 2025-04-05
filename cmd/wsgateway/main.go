package main

import (
	"os"

	"github.com/alikarimi999/shahboard/pkg/utils"
	"github.com/alikarimi999/shahboard/wsgateway"
)

func main() {
	file := os.Getenv("CONFIG_FILE")
	if file == "" {
		file = "./deploy/wsgateway/development/config.json"
	}

	cfg := &wsgateway.Config{}
	if err := utils.LoadConfigs(file, cfg); err != nil {
		panic(err)
	}

	app, err := wsgateway.SetupApplication(*cfg)
	if err != nil {
		panic(err)
	}

	app.Run()
}
