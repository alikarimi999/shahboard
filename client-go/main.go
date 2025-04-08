package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/alikarimi999/shahboard/client-go/bot"
	"github.com/alikarimi999/shahboard/client-go/stockfish"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

type Config struct {
	Server        string `json:"server"`
	StockfishPath string `json:"stockfish_path"`
	Bots          []Bot  `json:"bots"`
}

type Bot struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	// http.DefaultClient.Transport = &http.Transport{
	// 	TLSClientConfig: &tls.Config{
	// 		InsecureSkipVerify: true,
	// 	},
	// }

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
		b, err := bot.NewBot(bc.Email, bc.Password, cfg.Server, randSkill(), sp)
		if err != nil {
			fmt.Printf("bot '%s' error: %v\n", bc.Email, err)
			continue
		}
		handleBot(b)
	}

	select {}
}

func handleBot(b *bot.Bot) {
	go func() {
		for {
			// randSleep()
			ok, err := b.Login()
			if err != nil {
				fmt.Printf("bot '%s' login error: %v\n", b.Email(), err)
				continue
			}

			fmt.Printf("bot '%s' login success\n", b.Email())

			if !ok {
				if err := b.UpdateProfile(b.Email(), "avatar"); err != nil {
					fmt.Printf("bot '%s' update profile error: %v\n", b.Email(), err)
					continue
				}
			}

			if err := b.SetupWS(); err != nil {
				fmt.Printf("bot '%s' ws error: %v\n", b.Email(), err)
				continue
			}

			fmt.Printf("bot '%s' ws connected\n", b.Email())

			gameId, err := b.GetUserLiveGame(b.ID())
			if err != nil {
				fmt.Printf("bot '%s' get live game error: %v\n", b.Email(), err)
				continue
			}

			if !gameId.IsZero() {
				if err := b.Resume(gameId); err != nil {
					fmt.Printf("bot '%s' resume error: %v\n", b.Email(), err)
					continue
				}
				continue
			}

			e, err := b.FindMatch()
			if err != nil {
				fmt.Printf("bot '%s' find match error: %v\n", b.Email(), err)
				continue
			}

			if err := b.Create(e); err != nil {
				fmt.Printf("bot '%s' play error: %v\n", b.Email(), err)
				continue
			}

		}
	}()
}

func randSkill() int {
	return rand.Intn(15)
}

func randSleep() {
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
}
