package main

import (
	"os"

	"github.com/alikarimi999/shahboard/authservice"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

func main() {

	file := os.Getenv("CONFIG_FILE")
	if file == "" {
		file = "./deploy/auth/development/config.json"
	}

	cfg := &authservice.Config{}
	if err := utils.LoadConfigs(file, cfg); err != nil {
		panic(err)
	}

	app, err := authservice.SetupApplication(*cfg)
	if err != nil {
		panic(err)
	}

	if err := app.Run(); err != nil {
		panic(err)
	}

}
