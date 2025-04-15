package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/alikarimi999/shahboard/client-go/bot"
	"github.com/alikarimi999/shahboard/client-go/config"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

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

	cfg := &config.Config{}
	if err := utils.LoadConfigs("./config.json", cfg); err != nil {
		panic(err)
	}

	if !cfg.Local && cfg.Server == "" {
		panic("server is required")
	}

	// if cfg.StockfishPath == "" {
	// 	panic("stockfish path is required")
	// }

	bots := generateBots(0, cfg.BotsNum)

	// sp, err := stockfish.NewStockfish(cfg.StockfishPath)
	// if err != nil {
	// 	panic(err)
	// }

	for _, bc := range bots {
		go func() {
			for {
				startBot(cfg, bc)
			}
		}()
	}

	select {}
}

func startBot(cfg *config.Config, bc Bot) {
	randSleep(10)
	b, err := bot.NewBot(cfg, bc.Email, bc.Password, randSkill())
	if err != nil {
		fmt.Printf("bot '%s' error: %v\n", bc.Email, err)
		return
	}

	defer func() {
		b.Stop()
		fmt.Printf("bot '%s' stopped\n", b.Email())
	}()

	ok, err := b.Login()
	if err != nil {
		fmt.Printf("bot '%s' login error: %v\n", b.Email(), err)
		return
	}

	fmt.Printf("bot '%s' login success\n", b.Email())

	if !ok {
		go func() {
			randSleep(20)
			if err := b.UpdateProfile(strings.Split(b.Email(), "@")[0],
				fmt.Sprintf("https://robohash.org/%s.png", b.Email())); err != nil {
				fmt.Printf("bot '%s' update profile error: %v\n", b.Email(), err)
				return
			}
		}()
	}

	if err := b.SetupWS(); err != nil {
		fmt.Printf("bot '%s' ws error: %v\n", b.Email(), err)
		return
	}

	fmt.Printf("%d: bot '%s' ws connected\n", b.Email())

	go func() {
		b.RandomView()
	}()

	gameId, err := b.GetUserLiveGame(b.ID())
	if err != nil {
		fmt.Printf("bot '%s' get live game error: %v\n", b.Email(), err)
		return
	}

	if !gameId.IsZero() {
		if err := b.Resume(gameId); err != nil {
			fmt.Printf("bot '%s' resume error: %v\n", b.Email(), err)
			return
		}
		return
	}

	e, err := b.FindMatch()
	if err != nil {
		fmt.Printf("bot '%s' find match error: %v\n", b.Email(), err)
		return
	}

	if err := b.Create(e); err != nil {
		fmt.Printf("bot '%s' play error: %v\n", b.Email(), err)
		return
	}
}

func randSkill() int {
	return rand.Intn(15)
}

func generateBots(start, end int) []Bot {
	var bots []Bot
	for i := start; i < end; i++ {
		bots = append(bots, Bot{
			Email:    fmt.Sprintf("bot%d@gmail.com", i),
			Password: fmt.Sprintf("bot%d_password", i),
		})
	}
	return bots
}

func randSleep(max int) {
	time.Sleep(time.Duration((rand.Intn(max))+5) * time.Second)
}
