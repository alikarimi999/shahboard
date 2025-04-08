package bot

import (
	"sync"

	"github.com/alikarimi999/shahboard/client-go/stockfish"

	"github.com/alikarimi999/shahboard/types"
)

type Bot struct {
	id       types.ObjectId
	email    string
	password string
	skill    int
	jwtToken string
	ws       *webSocket
	sp       *stockfish.Stockfish
	ps       *PubSub

	g *game

	url string

	once   sync.Once
	stopCh chan struct{}
}

func NewBot(email, password, baseURL string, skill int, sp *stockfish.Stockfish) (*Bot, error) {
	b := &Bot{
		email:    email,
		password: password,
		skill:    skill,
		sp:       sp,
		ps:       NewPubSub(),
		url:      baseURL,
		stopCh:   make(chan struct{}),
	}

	return b, nil
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

		if b.ws != nil {
			b.ws.stop()
		}

		if b.g != nil {
			b.g.stop()
		}
	})
}
