package gameservice

import (
	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/gameservice/delivery/http"
	game "github.com/alikarimi999/shahboard/gameservice/service"
)

type Config struct {
	GameService game.Config  `json:"game_service"`
	Kafka       kafka.Config `json:"kafka"`
	Redis       RedisConfg   `json:"redis"`
	Log         LogConfig    `json:"log"`
	Http        http.Config  `json:"http"`
}

type RedisConfg struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type LogConfig struct {
	File    string `json:"file"`
	Verbose bool   `json:"verbose"`
}
