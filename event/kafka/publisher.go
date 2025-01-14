package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/alikarimi999/shahboard/event"
)

type kafkaPublisher struct {
	p sarama.SyncProducer
}

func newKafkaPublisher(brokers []string) (*kafkaPublisher, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &kafkaPublisher{p: producer}, nil
}

func (kp *kafkaPublisher) Publish(data ...event.Event) error {
	var errs []error
	for _, e := range data {
		msg := &sarama.ProducerMessage{
			Topic: e.GetTopic().Domain().String(),
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte(headerAction),
					Value: []byte(e.GetAction().String()),
				},
			},
			Key:   sarama.ByteEncoder(e.GetTopic().String()),
			Value: sarama.ByteEncoder(e.Encode()),
		}
		if _, _, err := kp.p.SendMessage(msg); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors occurred: %v", errs)
	}
	return nil
}

func (kp *kafkaPublisher) Close() error {
	return kp.p.Close()
}
