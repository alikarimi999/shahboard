package bot

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
	"github.com/notnil/chess"
)

var defaultNotation = chess.AlgebraicNotation{}

type game struct {
	b          *Bot
	id         types.ObjectId
	color      types.Color
	opponentId types.ObjectId
	board      *chess.Game
	subs       map[Topic]*Subscription

	moveDelay time.Duration

	once   sync.Once
	stopCh chan struct{}
}

func (b *Bot) Resume(gameId types.ObjectId) error {
	if b.g != nil {
		return fmt.Errorf("game already exists")
	}

	ms := b.Subscribe(Topic(ws.MsgTypeResumeGame))

	if err := b.SendWsMessage(ws.Msg{
		MsgBase: ws.MsgBase{
			Type:      ws.MsgTypeResumeGame,
			Timestamp: time.Now().Unix(),
		},
		Data: ws.DataResumeGameRequest{
			GameId: gameId,
		}.Encode(),
	}); err != nil {
		return err
	}

	e := <-ms.Consume()
	ms.Unsubscribe()
	fmt.Println("here >>> ", e)
	msg := ws.DataResumeGameResponse{}
	if err := json.Unmarshal(e.Data.(ws.Msg).Data, &msg); err != nil {
		return err
	}

	f, err := chess.PGN(strings.NewReader(msg.Pgn))
	if err != nil {
		return err
	}

	g := &game{
		b:  b,
		id: gameId,

		board:     chess.NewGame(f),
		subs:      make(map[Topic]*Subscription),
		moveDelay: 5 * time.Second,
		stopCh:    make(chan struct{}),
	}

	g.addBasicSubs(b.Subscribe)

	white := g.board.GetTagPair("w")
	black := g.board.GetTagPair("b")
	if white == nil || black == nil {
		return fmt.Errorf("game pgn is invalid")
	}

	if white.Value == b.ID().String() {
		g.color = types.ColorWhite
		g.opponentId = types.ObjectId(black.Value)
	} else {
		g.color = types.ColorBlack
		g.opponentId = types.ObjectId(white.Value)
	}

	b.g = g

	return g.run()
}

func (b *Bot) Create(ec event.EventUsersMatchCreated) error {
	if b.g != nil {
		return fmt.Errorf("game already exists")
	}

	ms := b.Subscribe(Topic(ws.MsgTypeGameCreate))

	if err := b.SendWsMessage(ws.Msg{
		MsgBase: ws.MsgBase{
			Type:      ws.MsgTypeFindMatch,
			Timestamp: time.Now().Unix(),
		},
		Data: ec.Encode(),
	}); err != nil {
		return err
	}

	e := <-ms.Consume()
	ms.Unsubscribe()

	msg := event.EventGameCreated{}
	if err := json.Unmarshal(e.Data.(ws.Msg).Data, &msg); err != nil {
		return err
	}

	g := &game{
		b:         b,
		id:        msg.GameID,
		board:     chess.NewGame(),
		subs:      make(map[Topic]*Subscription),
		moveDelay: 5 * time.Second,
		stopCh:    make(chan struct{}),
	}

	g.addBasicSubs(b.Subscribe)

	if b.id == msg.Player1.ID {
		g.color = msg.Player1.Color
		g.opponentId = msg.Player2.ID
	} else {
		g.color = msg.Player2.Color
		g.opponentId = msg.Player1.ID
	}

	b.g = g

	return g.run()
}

func (g *game) addBasicSubs(subscribe func(Topic) *Subscription) {
	t := Topic(ws.MsgTypeMoveApproved)
	g.subs[t] = subscribe(t)
	t = Topic(ws.MsgTypeGameEnd)
	g.subs[t] = subscribe(t)
	t = Topic(ws.MsgTypeChatCreated)
	g.subs[t] = subscribe(t)
	t = Topic(ws.MsgTypeChatMsgApproved)
	g.subs[t] = subscribe(t)
	t = Topic(ws.MsgTypePlayerConnectionUpdated)
	g.subs[t] = subscribe(t)
}

func (g *game) run() error {

	fmt.Printf("game started between %s and %s\n", g.b.ID(), g.opponentId)

	// first move if use is white
	if g.color == types.ColorWhite {
		m := g.bestMove()
		if err := g.board.Move(m); err != nil {
			g.stop()
			return fmt.Errorf("move error: %v", err)
		}

		g.sendMove(m)
	}

	wg := sync.WaitGroup{}
	for t, s := range g.subs {
		wg.Add(1)
		go func(t Topic, s *Subscription) {
			defer wg.Done()
			for e := range s.Consume() {
				g.handleEvent(e)
			}
		}(t, s)
	}

	wg.Wait()
	return nil
}

func (g *game) stop() {
	g.once.Do(func() {
		for _, s := range g.subs {
			s.Unsubscribe()
		}
		close(g.stopCh)
		g.b.g = nil
	})
}

func (g *game) handleEvent(e Event) {
	switch e.Topic {
	case Topic(ws.MsgTypeMoveApproved):
		g.handleMoveApproved(e.Data.(ws.Msg).Data)

	case Topic(ws.MsgTypeGameEnd):
		g.handleGameEnd(e.Data.(ws.Msg).Data)

	}
}

func (g *game) handleGameEnd(data []byte) {
	var e event.EventGameEnded
	if err := json.Unmarshal(data, &e); err != nil {
		fmt.Printf("unmarshal game end error: %v\n", err)
		return
	}
	g.stop()
	fmt.Printf("game ended: %v\n", e)
}

func (g *game) handleMoveApproved(data []byte) {
	var e event.EventGameMoveApproved
	if err := json.Unmarshal(data, &e); err != nil {
		fmt.Printf("unmarshal move approved error: %v\n", err)
		return
	}

	if e.GameID != g.id || e.PlayerID != g.opponentId || e.Index != (len(g.board.Moves())-1) {
		return
	}

	if err := g.board.MoveStr(e.Move); err != nil {
		fmt.Printf("move error: %v\n", err)
		return
	}

	m := g.bestMove()
	if err := g.board.Move(m); err != nil {
		fmt.Printf("move error: %v\n", err)
		return
	}

	g.sendMove(m)
}

func (g *game) bestMove() *chess.Move {
	for {
		m, err := g.b.sp.BestMove(g.board.FEN(), g.b.skill)
		if err != nil {
			fmt.Printf("stockfish error: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		move, err := chess.UCINotation{}.Decode(g.board.Position(), strings.TrimSpace(m))
		if err != nil {
			fmt.Printf("decode move error: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		return move
	}
}

func (g *game) sendMove(m *chess.Move) {
	time.Sleep(g.moveDelay)
	for {
		if err := g.b.SendWsMessage(ws.Msg{
			MsgBase: ws.MsgBase{
				Type:      ws.MsgTypePlayerMove,
				Timestamp: time.Now().Unix(),
			},
			Data: ws.DataGamePlayerMoveRequest{
				EventGamePlayerMoved: event.EventGamePlayerMoved{
					GameID:   g.id,
					PlayerID: g.b.id,
					Move:     defaultNotation.Encode(g.board.Position(), m),
					Index:    len(g.board.Moves()),
				},
			}.Encode(),
		}); err != nil {
			fmt.Printf("send move error: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}
	}
}
