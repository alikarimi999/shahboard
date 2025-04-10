package bot

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/alikarimi999/shahboard/client-go/config"
	"github.com/alikarimi999/shahboard/client-go/stockfish"
	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/wsgateway/ws"

	"github.com/alikarimi999/shahboard/types"
)

type Bot struct {
	cfg      *config.Config
	id       types.ObjectId
	email    string
	password string
	skill    int
	jwtToken string
	ws       *webSocket
	sp       *stockfish.Stockfish
	ps       *PubSub

	subs map[Topic]*Subscription

	g      *game
	vm     *viewManager
	once   sync.Once
	stopCh chan struct{}
}

func NewBot(cfg *config.Config, email, password string,
	skill int, sp *stockfish.Stockfish) (*Bot, error) {
	b := &Bot{
		cfg:      cfg,
		email:    email,
		password: password,
		skill:    skill,
		sp:       sp,
		subs:     make(map[Topic]*Subscription),
		ps:       NewPubSub(),
		vm:       NewViewManager(email),
		stopCh:   make(chan struct{}),
	}

	t := Topic(ws.MsgTypeGameEnd)
	b.subs[t] = b.Subscribe(t)
	t = Topic(ws.MsgTypeViewGame)
	b.subs[t] = b.Subscribe(t)

	b.run()
	return b, nil
}

func (b *Bot) run() {
	for t, s := range b.subs {
		go func(t Topic, s *Subscription) {
			for e := range s.Consume() {
				b.handleEvent(e)
			}
		}(t, s)
	}
}

func (b *Bot) ID() types.ObjectId {
	return b.id
}

func (b *Bot) Email() string {
	return b.email
}

func (b *Bot) Stop() {
	b.once.Do(func() {
		close(b.stopCh)

		for _, s := range b.subs {
			s.Unsubscribe()
		}

		if b.ws != nil {
			b.ws.stop()
		}

		if b.g != nil {
			b.g.stop()
		}
	})
}

func (b *Bot) handleEvent(e Event) {
	switch e.Topic {
	case Topic(ws.MsgTypeViewGame):
		b.handleGameView(e.Data.(*ws.Msg).Data)
	case Topic(ws.MsgTypeGameEnd):
		b.handleGameEnd(e.Data.(*ws.Msg).Data)

	}
}

func (b *Bot) handleGameView(data []byte) {
	var msg ws.DataGameViewResponse
	if err := json.Unmarshal(data, &msg); err != nil {
		fmt.Printf("unmarshal game view error: %v\n", err)
		return
	}

	b.vm.add(msg.GameId)
}

func (b *Bot) handleGameEnd(data []byte) {
	var e event.EventGameEnded
	if err := json.Unmarshal(data, &e); err != nil {
		fmt.Printf("unmarshal game end error: %v\n", err)
		return
	}

	b.vm.remove(e.GameID)
}
