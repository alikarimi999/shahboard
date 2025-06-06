package wsgateway

import (
	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/pkg/grpc"
	wsgrcp "github.com/alikarimi999/shahboard/wsgateway/grpc"

	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
)

type Config struct {
	Ws           ws.WsConfigs        `json:"ws"`
	Kafka        kafka.Config        `json:"kafka"`
	Log          LogConfig           `json:"log"`
	Http         HttpConfig          `json:"http"`
	Redis        RedisConfig         `json:"redis"`
	JwtValidator jwt.ValidatorConfig `json:"jwt_validator"`
	Grpc         wsgrcp.Config       `json:"grpc"`
	GameService  grpc.Config         `json:"game_service_grpc"`
}

type LogConfig struct {
	File    string `json:"file"`
	Verbose bool   `json:"verbose"`
}

type HttpConfig struct {
	Port int `json:"port"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}
