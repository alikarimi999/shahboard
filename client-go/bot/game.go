package bot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
	"github.com/notnil/chess"
)

var defaultNotation = chess.AlgebraicNotation{}
var defaultDelay int = 5
var defaultMsgLen int = 5

type game struct {
	b          *Bot
	id         types.ObjectId
	color      types.Color
	opponentId types.ObjectId
	board      *chess.Game
	subs       map[Topic]*Subscription

	defaultDelay int

	gamesCount *atomic.Int32
	once       sync.Once
	stopCh     chan struct{}
}

func (b *Bot) Resume(gameId types.ObjectId, counter *atomic.Int32) error {
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

	msg := ws.DataResumeGameResponse{}
	msgData, ok := e.Data.(*ws.Msg)
	if !ok {
		return fmt.Errorf("unexpected message type: %T", e.Data)
	}
	if err := json.Unmarshal(msgData.Data, &msg); err != nil {
		return err
	}

	f, err := chess.PGN(strings.NewReader(msg.Pgn))
	if err != nil {
		return err
	}

	g := &game{
		b:  b,
		id: gameId,

		board:        chess.NewGame(f),
		subs:         make(map[Topic]*Subscription),
		gamesCount:   counter,
		defaultDelay: defaultDelay,
		stopCh:       make(chan struct{}),
	}

	if g.board.Outcome() != chess.NoOutcome {
		return fmt.Errorf("game is already finished")
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

func (b *Bot) Create(ec event.EventUsersMatchCreated, counter *atomic.Int32) error {
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
	msgData, ok := e.Data.(*ws.Msg)
	if !ok {
		return fmt.Errorf("unexpected message type: %T", e.Data)
	}
	if err := json.Unmarshal(msgData.Data, &msg); err != nil {
		return err
	}

	g := &game{
		b:            b,
		id:           msg.GameID,
		board:        chess.NewGame(),
		subs:         make(map[Topic]*Subscription),
		gamesCount:   counter,
		defaultDelay: defaultDelay,
		stopCh:       make(chan struct{}),
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
	t = Topic(ws.MsgTypePlayerJoined)
	g.subs[t] = subscribe(t)
	t = Topic(ws.MsgTypePlayerLeft)
	g.subs[t] = subscribe(t)
}

func (g *game) run() error {
	endSleep := time.Duration(0)
	defer func() {
		time.Sleep(endSleep)
	}()

	// this is a stopChance function to test system reaction to player's disconnection
	go func() {
		if chance(g.b.cfg.StopChance) {
			endSleep = time.Duration(rand.Intn(120)+30) * time.Second
			randSleep(120)
			g.b.Stop()
			fmt.Printf("bot '%s' disconnected from game '%s' randomly\n", g.b.Email(), g.id)
		}
	}()

	g.chatGenerator()
	g.resigner()

	// first move
	if types.Color(g.board.Position().Turn()) == g.color {
		g.gamesCount.Add(1)
		fmt.Printf("%d: game started between %s and %s\n", g.gamesCount.Load(), g.b.ID(), g.opponentId)

		randSleep(g.defaultDelay)
		// fmt.Println(g.b.Email(), g.board.Position().Turn(), g.color)
		m, err := g.randMove()
		if err != nil {
			g.stop()
			return err
		}

		if err := g.board.MoveStr(m); err != nil {
			g.stop()
			return fmt.Errorf("move error: %v", err)
		}

		if err := g.sendMove(m); err != nil {
			g.stop()
			return err
		}
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
		g.gamesCount.Add(-1)
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
		g.handleMoveApproved(e.Data.(*ws.Msg).Data)
	case Topic(ws.MsgTypeChatCreated):
		// g.handleChatCreated(e.Data.(*ws.Msg).Data)
	case Topic(ws.MsgTypeGameEnd):
		g.handleGameEnd(e.Data.(*ws.Msg).Data)

	}
}

func (g *game) handleChatCreated(data []byte) {
	fmt.Println("received chat created event")
	var e event.EventGameChatCreated
	if err := json.Unmarshal(data, &e); err != nil {
		fmt.Printf("unmarshal chat created error: %v\n", err)
		return
	}

	if g.id != e.GameID {
		return
	}

	fmt.Printf("chat enabled: %v\n", e)
	s, ok := g.subs[Topic(ws.MsgTypeChatCreated)]
	if ok {
		s.Unsubscribe()
	}

	g.chatGenerator()
}

func (g *game) handleGameEnd(data []byte) {
	var e event.EventGameEnded
	if err := json.Unmarshal(data, &e); err != nil {
		fmt.Printf("unmarshal game end error: %v\n", err)
		return
	}

	if e.GameID == g.id {
		g.stop()
		return
	}
}

func (g *game) handleMoveApproved(data []byte) {
	var e event.EventGameMoveApproved
	if err := json.Unmarshal(data, &e); err != nil {
		fmt.Printf("unmarshal move approved error: %v\n", err)
		return
	}

	if e.GameID != g.id || e.PlayerID != g.opponentId || e.Index-1 != len(g.board.Moves()) {
		return
	}
	// fmt.Printf("bot %s received move %s with index %d, current index %d\n",
	// g.b.ID(), e.Move, e.Index, len(g.board.Moves()))

	if err := g.board.MoveStr(e.Move); err != nil {
		fmt.Printf("move error: %v\n", err)
		return
	}

	if g.board.Outcome() != chess.NoOutcome {
		fmt.Printf("game %s ended\n", g.id)
		return
	}

	m, err := g.randMove()
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := g.board.MoveStr(m); err != nil {
		fmt.Printf("move error: %v\n", err)
		return
	}

	if err := g.sendMove(m); err != nil {
		fmt.Println(err)
		return
	}
}

func (g *game) bestMove() (*chess.Move, error) {
	count := 0
	for {
		count++
		m, err := g.b.sp.BestMove(g.board.FEN(), g.b.skill)
		if err != nil {
			fmt.Printf("stockfish error: %v\n", err)
			if count == 3 {
				return nil, fmt.Errorf("stockfish error: %v", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		move, err := chess.UCINotation{}.Decode(g.board.Position(), strings.TrimSpace(m))
		if err != nil {
			fmt.Printf("decode move error: %v\n", err)
			if count == 3 {
				return nil, fmt.Errorf("decode move error: %v", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		return move, nil
	}
}

func (g *game) randMove() (string, error) {

	vm := g.board.ValidMoves()
	for _, m := range vm {
		ms := defaultNotation.Encode(g.board.Position(), m)
		if strings.Contains(ms, "x") {
			return ms, nil
		}
	}
	return defaultNotation.Encode(g.board.Position(), vm[rand.Intn(len(vm))]), nil
}

func (g *game) sendMove(m string) error {
	count := 0
	for {
		randSleep(g.defaultDelay)
		count++
		index := len(g.board.Moves())
		// fmt.Printf("bot %s send move %s index %d\n", g.b.ID(), m, index)
		if err := g.b.SendWsMessage(ws.Msg{
			MsgBase: ws.MsgBase{
				Type:      ws.MsgTypePlayerMove,
				Timestamp: time.Now().Unix(),
			},
			Data: ws.DataGamePlayerMoveRequest{
				EventGamePlayerMoved: event.EventGamePlayerMoved{
					GameID:   g.id,
					PlayerID: g.b.id,
					Move:     m,
					Index:    index,
				},
			}.Encode(),
		}); err != nil {
			fmt.Printf("send move error: %v\n", err)
			if count == 3 {
				return fmt.Errorf("send move error: %v", err)
			}
			continue
		}
		return nil
	}
}

func (g *game) sendChat(msg string) error {
	count := 0
	for {
		randSleep(g.defaultDelay)
		count++
		t := time.Now().Unix()
		if err := g.b.SendWsMessage(ws.Msg{
			MsgBase: ws.MsgBase{
				Type:      ws.MsgTypeChatMsgSend,
				Timestamp: t,
			},
			Data: ws.DataGameChatMsgSend{
				EventGameChatMsgeSent: event.EventGameChatMsgeSent{
					SenderID:  g.b.ID(),
					GameID:    g.id,
					Content:   msg,
					Timestamp: t,
				},
			}.Encode(),
		}); err != nil {
			fmt.Printf("send chat error: %v\n", err)
			if count == 3 {
				return fmt.Errorf("send chat error: %v", err)
			}
			continue
		}
		return nil
	}
}

func randSleep(max int) {
	time.Sleep(time.Duration((rand.Intn(max))+5) * time.Second)
}
