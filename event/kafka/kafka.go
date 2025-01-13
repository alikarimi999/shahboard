package kafka

import (
	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
)

type Config struct {
	Brokers []string `json:"brokers"`
	GroupID string   `json:"group_id"`
}

func NewKafkaPublisherAndSubscriber(cfg Config, l log.Logger) (event.Publisher, event.Subscriber, error) {
	p, err := newKafkaPublisher(cfg.Brokers)
	if err != nil {
		return nil, nil, err
	}
	s, err := newKafkaSubscriber(cfg.Brokers, cfg.GroupID, l)
	if err != nil {
		return nil, nil, err
	}

	return p, s, nil
}
