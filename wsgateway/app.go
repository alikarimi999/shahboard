package wsgateway

import (
	"context"
	"fmt"

	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/pkg/grpc"
	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/pkg/middleware"
	wsgrpc "github.com/alikarimi999/shahboard/wsgateway/grpc"
	"github.com/alikarimi999/shahboard/wsgateway/services/game"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type application struct {
	cfg        Config
	server     *ws.Server
	e          *gin.Engine
	grpcServer *wsgrpc.Server
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

	// gin.SetMode(gin.ReleaseMode)
	// e := gin.Default()

	e := gin.New()
	e.Use(gin.Recovery(), middleware.Cors())

	c := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err = c.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	client, err := grpc.NewClient(cfg.GameService, nil)
	if err != nil {
		return nil, err
	}

	server, err := ws.NewServer(e, s, p, game.NewService(client), &cfg.Ws, c, v, l)
	if err != nil {
		return nil, err
	}

	gs, err := wsgrpc.NewServer(cfg.Grpc, server)
	if err != nil {
		return nil, err
	}

	return &application{
		cfg:        cfg,
		server:     server,
		e:          e,
		grpcServer: gs,
	}, nil
}

func (a *application) Run() error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.grpcServer.Run()
	}()

	go func() {
		errCh <- a.e.Run(fmt.Sprintf(":%d", a.cfg.Http.Port))
	}()

	return <-errCh
}
