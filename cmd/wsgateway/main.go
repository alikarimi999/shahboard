package main

import (
	"os"

	"github.com/alikarimi999/shahboard/pkg/utils"
	"github.com/alikarimi999/shahboard/wsgateway"
)

func main() {
	cfg := &wsgateway.Config{}
	if err := utils.LoadConfigs(os.Getenv("CONFIG_FILE"), cfg); err != nil {
		panic(err)
	}

	app, err := wsgateway.SetupApplication(*cfg)
	if err != nil {
		panic(err)
	}

	app.Run()
}
