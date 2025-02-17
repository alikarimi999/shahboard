package main

import (
	"fmt"

	"github.com/alikarimi999/shahboard/chatservice"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

func main() {
	cfg := &chatservice.Config{}
	if err := utils.LoadConfigs("chat", true, cfg); err != nil {
		panic(err)
	}

	_, err := chatservice.SetupApplication(*cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println("Chat service is running...")
	select {}

}
