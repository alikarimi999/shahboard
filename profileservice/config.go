package profileservice

import (
	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/pkg/postgres"
	"github.com/alikarimi999/shahboard/pkg/router"
	"github.com/alikarimi999/shahboard/profileservice/delivery/grpc"
	"github.com/alikarimi999/shahboard/profileservice/service/rating"
	"github.com/alikarimi999/shahboard/profileservice/service/user"
)

type Config struct {
	User         user.Config         `json:"user_service"`
	Rating       rating.Config       `json:"rating_service"`
	Kafka        kafka.Config        `json:"kafka"`
	Log          LogConfig           `json:"log"`
	JwtValidator jwt.ValidatorConfig `json:"jwt_validator"`
	UsersDB      postgres.Config     `json:"users_db"`
	RatingDB     postgres.Config     `json:"rating_db"`
	Redis        RedisConfig         `json:"redis"`
	Http         router.Config       `json:"http"`
	Grpc         grpc.Config         `json:"grpc_server"`
}

type LogConfig struct {
	File    string `json:"file"`
	Verbose bool   `json:"verbose"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}
