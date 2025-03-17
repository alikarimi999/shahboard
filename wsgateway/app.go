package wsgateway

import (
	"context"
	"fmt"

	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/pkg/middleware"
	"github.com/alikarimi999/shahboard/wsgateway/services/game"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type application struct {
	server *ws.Server
	e      *gin.Engine
	cfg    Config
}

func SetupApplication(cfg Config) (*application, error) {
	l := log.NewLogger(cfg.Log.File, cfg.Log.Verbose)

	v, err := jwt.NewValidator(cfg.JwtValidator)
	if err != nil {
		return nil, err
	}

	p, s, err := kafka.NewKafkaPublisherAndSubscriber(cfg.Kafka, l)
	if err != nil {
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)
	e := gin.Default()
	e.Use(middleware.Cors())

	c := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err = c.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	game, err := game.NewService(cfg.GameService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	server, err := ws.NewServer(e, s, p, game, nil, c, v, l)
	if err != nil {
		return nil, err
	}

	return &application{
		server: server,
		e:      e,
		cfg:    cfg,
	}, nil
}

func (a *application) Start() error {
	return a.e.Run(fmt.Sprintf(":%s", a.cfg.Http.Port))
}
