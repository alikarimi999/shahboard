package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
)

// kafkaSubscriber Implement the event.Subscriber interface.
// It handles consuming events from Kafka topics and broadcasting them to subscribers.
type kafkaSubscriber struct {
	brokers []string
	groupId string

	g                   *consumerGroup
	subscriptionManager *subscriptionManager // Subscription manager for managing subscriptions.

	closeSignal chan struct{}  // Channel to signal shutdown.
	wg          sync.WaitGroup // WaitGroup to synchronize goroutines.

	l log.Logger
}

func newKafkaSubscriber(brokers []string, groupID string, l log.Logger) (*kafkaSubscriber, error) {
	g, err := newConsumerGroup(brokers, groupID, l)
	if err != nil {
		l.Error(fmt.Sprintf("Failed to create a new consumer group: %v", err))
		return nil, err
	}

	kc := &kafkaSubscriber{
		brokers: brokers,
		groupId: groupID,

		g:                   g,
		subscriptionManager: newSubscriptionManager(),
		closeSignal:         make(chan struct{}),
		l:                   l,
	}

	l.Info("kafka subscriber created successfully")
	return kc, nil
}

func (kc *kafkaSubscriber) Subscribe(topic event.Topic) event.Subscription {
	sub := newFeedSub(topic, kc)
	kc.subscriptionManager.addSub(sub)
	kc.g.consume(kc, string(topic.Domain()))
	return sub
}

// Closes the consumer and waits for all goroutines to finish.
func (kc *kafkaSubscriber) Close() error {
	kc.l.Info("closing kafka subscriber")
	close(kc.closeSignal) // Signals the shutdown of the consumer.
	kc.g.close()
	kc.wg.Wait() // Waits for all goroutines to finish.
	kc.l.Info("kafka subscriber closed")
	return nil
}

type consumerGroupHandler struct {
	kc     *kafkaSubscriber
	topics []string
	l      log.Logger
}

func newConsumerGroupHandler(kc *kafkaSubscriber, topics []string, l log.Logger) sarama.ConsumerGroupHandler {
	return &consumerGroupHandler{
		kc:     kc,
		topics: topics,
		l:      l,
	}
}

// No-op methods for the Sarama consumer group interface.
func (ch *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	ch.l.Debug(fmt.Sprintf("consumer group handler setup for topics: %v", ch.topics))
	return nil
}
func (ch *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	ch.l.Debug(fmt.Sprintf("consumer group handler cleanup for topics %v", ch.topics))
	return nil
}

// Consumes messages from Kafka, decodes events, and broadcasts them to subscribers.
func (
	ch *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ch.l.Debug(fmt.Sprintf("starting to consume messages for topics %v", ch.topics))
	for message := range claim.Messages() {
		var action string
		for _, header := range message.Headers {
			if string(header.Key) == headerAction {
				action = string(header.Value)
			}
		}

		if action == "" {
			continue
		}

		e, err := decodeEvent(message.Topic, action, message.Value)
		if err == nil {
			ch.kc.subscriptionManager.send(e) // Sends the decoded event to all subscribers.
		}

	}
	return nil
}

type index uint64

// subscriptionManager handles subscriptions for topics and broadcasts events to subscribers.
type subscriptionManager struct {
	subs map[event.Domain]map[index]*feedSub
	mu   sync.Mutex
}

func newSubscriptionManager() *subscriptionManager {
	return &subscriptionManager{
		subs: make(map[event.Domain]map[index]*feedSub),
	}
}

// Adds a new subscriber to a topic.
func (sm *subscriptionManager) addSub(sub *feedSub) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exists := sm.subs[sub.topic.Domain()]; !exists {
		sm.subs[sub.topic.Domain()] = make(map[index]*feedSub)
	}
	sub.i = index(len(sm.subs[sub.topic.Domain()])) // Sets a unique index for the subscriber.
	sm.subs[sub.topic.Domain()][sub.i] = sub
}

// Removes a subscriber from a topic.
func (sm *subscriptionManager) removeSub(sub *feedSub) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.subs[sub.topic.Domain()], sub.i)
}

// Sends an event to all subscribers of the event's topic.
func (sm *subscriptionManager) send(events ...event.Event) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	// subscribers := make([]chan event.Event, 0)
	for _, e := range events {
		if subs, exists := sm.subs[e.GetTopic().Domain()]; exists {
			for _, sub := range subs {
				if e.GetTopic().Match(sub.topic) {
					select {
					case sub.ch <- e:
					default:
						fmt.Println("event dropped: ", e.GetTopic().String())
					}

					// subscribers = append(subscribers, sub.ch)
				}
			}
		}
	}

	// for _, ch := range subscribers {
	// 	for _, e := range events {
	// 		select {
	// 		case ch <- e:
	// 		default:
	// 			fmt.Println("event dropped: ", e.GetTopic().String())
	// 		}

	// 	}
	// }
}

// feedSub Implements the event.Subscription interface.
// It represents a subscription to a specific topic, including the channel for events and errors.
type feedSub struct {
	i     index // Unique index for the subscriber.
	topic event.Topic
	kc    *kafkaSubscriber

	ch  chan event.Event // Channel for receiving events.
	err chan error       // Channel for receiving errors.
}

func newFeedSub(topic event.Topic, kc *kafkaSubscriber) *feedSub {
	return &feedSub{
		topic: topic,
		kc:    kc,

		ch:  make(chan event.Event, 1000),
		err: make(chan error, 1000),
	}
}

func (s *feedSub) Topic() event.Topic { return s.topic }

func (s *feedSub) Event() <-chan event.Event { return s.ch }

func (s *feedSub) Err() <-chan error { return s.err }

// Unsubscribes from the topic and removes the topic from the topic manager if there are no more subscribers.
func (s *feedSub) Unsubscribe() {
	s.kc.subscriptionManager.removeSub(s) // Removes the subscription.
	close(s.ch)
	close(s.err)
}

func decodeEvent(domain, action string, data []byte) (event.Event, error) {
	var e event.Event
	switch domain {
	case event.DomainGame:
		switch event.Action(action) {
		case event.ActionCreated:
			e = &event.EventGameCreated{}
		case event.ActionGamePlayerMoved:
			e = &event.EventGamePlayerMoved{}
		case event.ActionGameMoveApprove:
			e = &event.EventGameMoveApproved{}
		case event.ActionGamePlayerLeft:
			e = &event.EventGamePlayerLeft{}
		case event.ActionGamePlayerConnectionUpdated:
			e = &event.EventGamePlayerConnectionUpdated{}
		case event.ActionGamePlayerSelectSquare:
			e = &event.EventGamePlayerSelectSquare{}
		case event.ActionEnded:
			e = &event.EventGameEnded{}
		default:
			return nil, fmt.Errorf("unknown event type for topic: %s", domain)
		}
	case event.DomainMatch:
		switch event.Action(action) {
		case event.ActionUsersMatched:
			e = &event.EventUsersMatched{}
		default:
			return nil, fmt.Errorf("unknown event type for topic: %s", domain)
		}
	default:
		return nil, fmt.Errorf("unknown event type for topic: %s", domain)
	}

	// Decode the JSON data into the event struct
	if err := json.Unmarshal(data, e); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to unmarshal event data: %v", err)
	}

	return e, nil
}

type consumerGroup struct {
	mu sync.Mutex

	isRunnig bool
	ctx      context.Context
	cancel   context.CancelFunc

	brokers []string
	groupId string
	cfg     *sarama.Config

	g sarama.ConsumerGroup

	td []string
	wg sync.WaitGroup
	l  log.Logger
}

func newConsumerGroup(brokers []string, groupID string, l log.Logger) (*consumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	cg, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &consumerGroup{

		brokers: brokers,
		groupId: groupID,
		cfg:     config,

		g: cg,
		l: l,
	}, nil
}

func (cg *consumerGroup) consume(kc *kafkaSubscriber, domain string) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	for _, t := range cg.td {
		if t == domain {
			return
		}
	}

	if !cg.isRunnig {
		ctx, cancel := context.WithCancel(context.Background())
		cg.ctx = ctx
		cg.cancel = cancel
		cg.isRunnig = true
		cg.td = append(cg.td, domain)
		cg.wg.Add(1)
		go func() {
			defer cg.wg.Done()
			if err := cg.g.Consume(ctx, cg.td, newConsumerGroupHandler(kc, cg.td, cg.l)); err != nil {
				if err.Error() != context.Canceled.Error() {
					cg.l.Error(fmt.Sprintf("Error consuming domains %v: %v", cg.td, err))
				}
			}
		}()

		return
	}
	cg.cancel()
	cg.wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	cg.ctx = ctx
	cg.cancel = cancel
	cg.td = append(cg.td, domain)

	cg.l.Info("restarting consumer group")
	for {
		newCg, err := sarama.NewConsumerGroup(cg.brokers, cg.groupId, cg.cfg)
		if err != nil {
			cg.l.Error(fmt.Sprintf("failed to create a new consumer group during restart: %v", err))
			time.Sleep(time.Second)
			continue
		}

		cg.g = newCg
		break
	}
	cg.l.Info("restarted consumer group")

	cg.wg.Add(1)
	go func() {
		defer cg.wg.Done()
		if err := cg.g.Consume(ctx, cg.td, newConsumerGroupHandler(kc, cg.td, cg.l)); err != nil {
			cg.l.Error(fmt.Sprintf("Error consuming domains %v: %v", cg.td, err))
		}
	}()

}

func (cg *consumerGroup) close() {
	cg.g.Close()
	cg.wg.Wait()
}
