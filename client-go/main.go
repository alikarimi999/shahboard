package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alikarimi999/shahboard/client-go/bot"
	"github.com/alikarimi999/shahboard/client-go/stockfish"
	"github.com/alikarimi999/shahboard/pkg/utils"
)

type Config struct {
	Server        string `json:"server"`
	StockfishPath string `json:"stockfish_path"`
	BotsNum       int    `json:"bots_num"`
}

type Bot struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var wsCounter atomic.Int32
var gameCounter *atomic.Int32 = &atomic.Int32{}

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

	bots := generateBots(cfg.BotsNum)

	sp, err := stockfish.NewStockfish(cfg.StockfishPath)
	if err != nil {
		panic(err)
	}

	for _, bc := range bots {
		go func() {
			for {
				startBot(bc, cfg.Server, sp)
			}
		}()
	}

	select {}
}

func startBot(bc Bot, url string, sp *stockfish.Stockfish) {
	randSleep(10)
	b, err := bot.NewBot(bc.Email, bc.Password, url, randSkill(), sp)
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
		if err := b.UpdateProfile(strings.Split(b.Email(), "@")[0],
			fmt.Sprintf("https://robohash.org/%s.png", b.Email())); err != nil {
			fmt.Printf("bot '%s' update profile error: %v\n", b.Email(), err)
			return
		}
	}

	if err := b.SetupWS(); err != nil {
		fmt.Printf("bot '%s' ws error: %v\n", b.Email(), err)
		return
	}
	wsCounter.Add(1)
	fmt.Printf("%d: bot '%s' ws connected\n", wsCounter.Load(), b.Email())

	gameId, err := b.GetUserLiveGame(b.ID())
	if err != nil {
		fmt.Printf("bot '%s' get live game error: %v\n", b.Email(), err)
		return
	}

	if !gameId.IsZero() {
		if err := b.Resume(gameId, gameCounter); err != nil {
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

	if err := b.Create(e, gameCounter); err != nil {
		fmt.Printf("bot '%s' play error: %v\n", b.Email(), err)
		return
	}
}

func randSkill() int {
	return rand.Intn(15)
}

func generateBots(n int) []Bot {
	var bots []Bot
	for i := 0; i < n; i++ {
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
