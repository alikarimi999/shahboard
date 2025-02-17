package chatservice

import (
	chat "github.com/alikarimi999/shahboard/chatservice/service.go"
	"github.com/alikarimi999/shahboard/event/kafka"
)

type Config struct {
	Chat  chat.Config  `json:"chat_service"`
	Kafka kafka.Config `json:"kafka"`
	Redis RedisConfg   `json:"redis"`
	Log   LogConfig    `json:"log"`
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
