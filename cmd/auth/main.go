package main

import (
	"os"

	"github.com/alikarimi999/shahboard/authservice"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

func main() {

	cfg := &authservice.Config{}
	if err := utils.LoadConfigs(os.Getenv("CONFIG_FILE"), cfg); err != nil {
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
