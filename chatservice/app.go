package chatservice

import (
	"context"

	chat "github.com/alikarimi999/shahboard/chatservice/service.go"
	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/redis/go-redis/v9"
)

type application struct {
	ChatService *chat.Service
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

	a := &application{
		ChatService: chat.NewService(cfg.Chat, p, s, r, l),
	}

	return a, nil
}

func (a *application) Stop() {
	a.ChatService.Stop()
}
