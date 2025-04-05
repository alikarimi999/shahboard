package main

import (
	"os"

	"github.com/alikarimi999/shahboard/pkg/utils"
	"github.com/alikarimi999/shahboard/profileservice"
)

func main() {

	cfg := &profileservice.Config{}
	if err := utils.LoadConfigs(os.Getenv("CONFIG_FILE"), cfg); err != nil {
		panic(err)
	}

	app, err := profileservice.SetupApplication(*cfg)
	if err != nil {
		panic(err)
	}
	if err := app.Run(); err != nil {
		panic(err)
	}
}
