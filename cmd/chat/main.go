package main

import (
	"fmt"
	"os"

	"github.com/alikarimi999/shahboard/chatservice"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

func main() {
	file := os.Getenv("CONFIG_FILE")
	if file == "" {
		file = "./deploy/chat/development/config.json"
	}

	cfg := &chatservice.Config{}
	if err := utils.LoadConfigs(file, cfg); err != nil {
		panic(err)
	}

	_, err := chatservice.SetupApplication(*cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println("Chat service is running...")
	select {}

}
