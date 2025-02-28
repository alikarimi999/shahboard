package authservice

import (
	auth "github.com/alikarimi999/shahboard/authservice/service"
	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/pkg/postgres"
	"github.com/alikarimi999/shahboard/pkg/router"
)

type Config struct {
	Auth         auth.Config         `json:"auth_service"`
	Kafka        kafka.Config        `json:"kafka"`
	Log          LogConfig           `json:"log"`
	JwtGenerator jwt.GeneratorConfig `json:"jwt_generator"`
	PostgresDB   postgres.Config     `json:"postgres_db"`
	Http         router.Config       `json:"http"`
}

type LogConfig struct {
	File    string `json:"file"`
	Verbose bool   `json:"verbose"`
}
