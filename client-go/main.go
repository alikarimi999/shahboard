package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/alikarimi999/shahboard/client-go/bot"
	"github.com/alikarimi999/shahboard/client-go/stockfish"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

func main() {
	cfg := &Config{}
	if err := utils.LoadConfigs("./config.json", cfg); err != nil {
		panic(err)
	}

	if cfg.Server == "" {
		panic("server is required")
	}

	if cfg.StockfishPath == "" {
		panic("stockfish path is required")
	}

	if len(cfg.Bots) == 0 {
		panic("users are required")
	}

	sp, err := stockfish.NewStockfish(cfg.StockfishPath)
	if err != nil {
		panic(err)
	}

	for _, bc := range cfg.Bots {
		b, err := bot.NewBot(cfg.Server, bc.Email, bc.Password, randSkill(), sp)
		if err != nil {
			fmt.Printf("bot %s error: %v\n", bc.Email, err)
			continue
		}
		handleBot(b)
	}
}

func handleBot(b *bot.Bot) {
	go func() {
		for {
			randSleep()
			if err := b.Login(); err != nil {
				fmt.Printf("bot %s login error: %v\n", b.Email(), err)
				continue
			}

			if err := b.SetupWS(); err != nil {
				fmt.Printf("bot %s ws error: %v\n", b.Email(), err)
				continue
			}

			e, err := b.FindMatch()
			if err != nil {
				fmt.Printf("bot %s find match error: %v\n", b.Email(), err)
				continue
			}

			if err := b.Play(e); err != nil {
				fmt.Printf("bot %s play error: %v\n", b.Email(), err)
				continue
			}
		}
	}()
}

func randSkill() int {
	return rand.Intn(15)
}

func randSleep() {
	time.Sleep(time.Duration(rand.Intn(15)) * time.Second)
}
