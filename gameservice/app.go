package gameservice

import (
	"context"

	"github.com/alikarimi999/shahboard/event/kafka"
	game "github.com/alikarimi999/shahboard/gameservice/service"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/redis/go-redis/v9"
)

type application struct {
	gameService *game.Service
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

	return &application{
		gameService: gs,
	}, nil
}
