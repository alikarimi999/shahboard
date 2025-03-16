package matchservice

import (
	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/matchservice/delivery/game"
	"github.com/alikarimi999/shahboard/matchservice/delivery/http"
	match "github.com/alikarimi999/shahboard/matchservice/service"
	"github.com/alikarimi999/shahboard/pkg/jwt"
)

type Config struct {
	Match        match.Config        `json:"match_service"`
	Http         http.Config         `json:"http"`
	Kafka        kafka.Config        `json:"kafka"`
	Log          LogConfig           `json:"log"`
	JwtValidator jwt.ValidatorConfig `json:"jwt_validator"`
	GameService  game.Config         `json:"game_service_grpc"`
}

type LogConfig struct {
	File    string `json:"file"`
	Verbose bool   `json:"verbose"`
}
