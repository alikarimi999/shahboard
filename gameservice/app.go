package gameservice

import (
	"context"

	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/gameservice/delivery/grpc"
	"github.com/alikarimi999/shahboard/gameservice/delivery/http"
	game "github.com/alikarimi999/shahboard/gameservice/service"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/redis/go-redis/v9"
)

type application struct {
	GameService *game.Service
	Router      *http.Router
	Grpc        *grpc.Server
}

func SetupApplication(cfg Config) (*application, error) {
	l := log.NewLogger(cfg.Log.File, cfg.Log.Verbose)

	p, s, err := kafka.NewKafkaPublisherAndSubscriber(cfg.Kafka, l)
	if err != nil {
		return nil, err
	}

	r := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err = r.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	gs, err := game.NewGameService(cfg.GameService, r, p, s, l)
	if err != nil {
		return nil, err
	}

	router, err := http.NewRouter(cfg.Http, gs)
	if err != nil {
		return nil, err
	}

	grpcServer, err := grpc.NewServer(cfg.Grpc, gs)
	if err != nil {
		return nil, err
	}

	return &application{
		GameService: gs,
		Router:      router,
		Grpc:        grpcServer,
	}, nil
}

func (a *application) Run() error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.Grpc.Run()
	}()

	go func() {
		errCh <- a.Router.Run()
	}()

	return <-errCh
}
