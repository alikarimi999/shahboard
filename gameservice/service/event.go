package game

import (
	"context"
	"fmt"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/gameservice/entity"
)

func (gs *Service) subscribeEvents(topics ...event.Topic) {
	for _, topic := range topics {
		gs.subManager.addSub(gs.s.Subscribe(topic))
		gs.l.Debug(fmt.Sprintf("subscribed to topic: '%s'", topic))
	}
}

// handleEventPlayersMatched handles the event of two players being matched
// and creates a new game if both players do not have any active games.
// This function is idempotent and safe to call concurrently from multiple
// instances of the GameService.
func (gs *Service) handleEventPlayersMatched(d *event.EventUsersMatched) {

	// check if player is already in a game
	if gs.checkByPlayer(d.User1.ID) || gs.checkByPlayer(d.User2.ID) {
		gs.l.Debug("player is already in a game")
		return
	}

	// create a new game
	g := entity.NewGame(d.User1.ID, d.User2.ID, gs.cfg.DefaultGameSettings)

	// add the game to the cache
	if ok, err := gs.cache.addGame(context.Background(), g); err != nil {
		gs.l.Error(err.Error())
	} else if !ok {
		gs.l.Debug(fmt.Sprintf("game already exists in cache: '%s'", g.ID()))
		return
	}
	gs.l.Debug(fmt.Sprintf("added game to cache: '%s'", g.ID()))

	gm := newGameManager(gs, g)
	gs.addGame(gm)

	// subscribe to the game
	topic := event.TopicGame.WithResource(gm.ID().String())
	gm.addSub(gs.s.Subscribe(topic))
	gs.l.Debug(fmt.Sprintf("subscribed to topic: '%s'", topic))

	// publish the game created event
	if err := gs.p.Publish(event.EventGameCreated{
		ID:        g.ID(),
		Player1:   g.Player1(),
		Player2:   g.Player2(),
		Timestamp: time.Now().Unix(),
	}); err != nil {
		gs.l.Error(err.Error())
	}
	gs.l.Debug(fmt.Sprintf("published game created event: '%s'", g.ID()))

	gs.l.Debug(fmt.Sprintf("game '%s' created by player '%s' as '%s' and player '%s' as '%s'",
		g.ID(), g.Player1().ID, g.Player1().Color, g.Player2().ID, g.Player2().Color))

	// TODO: add to repository concurrency control
}

func (gs *Service) handleEventGamePlayerMoved(d *event.EventGamePlayerMoved) {
	g := gs.getGame(d.ID)
	// if game is not manging by this instance, do nothing
	if g == nil {
		return
	}

	if g.Turn().ID != d.PlayerID {
		gs.l.Debug(fmt.Sprintf("it's not player '%s' turn", d.PlayerID))
		return
	}

	if err := g.Move(d.Move); err != nil {
		gs.l.Debug(fmt.Sprintf("player '%s' made an invalid move '%s' on game '%s'", d.PlayerID, d.Move, d.ID))
		return
	}

	if g.Outcome() != entity.NoOutcome {
		g.Deactivate()
		if err := gs.cache.updateAndDeactivateGame(context.Background(), g.Game); err != nil {
			gs.l.Error(err.Error())
			return
		}
		gs.p.Publish(event.EventGameMoveApproved{
			PlayerID:  d.PlayerID,
			ID:        d.ID,
			Move:      d.Move,
			Timestamp: time.Now().Unix(),
		},
			event.EventGameEnded{
				ID:        g.ID(),
				Player1:   g.Player1().ID,
				Player2:   g.Player2().ID,
				Timestamp: time.Now().Unix(),
			})

		gs.l.Debug(fmt.Sprintf("published game move approved event: '%s'", g.ID()))
		gs.l.Debug(fmt.Sprintf("published game ended event: '%s'", g.ID()))

	} else {
		if err := gs.cache.updateGameMove(context.Background(), g.Game); err != nil {
			gs.l.Debug(err.Error())
			return
		}

		gs.l.Debug(fmt.Sprintf("player '%s' made move '%s' on game '%s'", d.PlayerID, d.Move, g.ID()))

		gs.p.Publish(event.EventGameMoveApproved{
			PlayerID:  d.PlayerID,
			ID:        d.ID,
			Move:      d.Move,
			Timestamp: time.Now().Unix(),
		})
		gs.l.Debug(fmt.Sprintf("published game move approved event: '%s'", g.ID()))
	}

	// TODO: think about how to update the database
}

func (gs *Service) handleEventGamePlayerLeft(d *event.EventGamePlayerLeft) {
	g := gs.getGame(d.ID)
	if g == nil {
		return
	}

	g.Deactivate()
	if err := gs.cache.updateAndDeactivateGame(context.Background(), g.Game); err != nil {
		gs.l.Error(err.Error())
		return
	}
	gs.removeGame(d.ID)
	gs.l.Debug(fmt.Sprintf("removed game: '%s'", d.ID))

	gs.p.Publish(event.EventGameEnded{
		ID:        d.ID,
		Player1:   g.Player1().ID,
		Player2:   g.Player2().ID,
		Timestamp: time.Now().Unix(),
	})
	gs.l.Debug(fmt.Sprintf("published game ended event: '%s'", d.ID))

}

func (gs *Service) handleEvents(e event.Event) {
	switch e.GetTopic().Domain() {
	case event.DomainGame:
		switch e.GetAction() {
		case event.ActionGamePlayerMoved:
			gs.handleEventGamePlayerMoved(e.(*event.EventGamePlayerMoved))
		case event.ActionGamePlayerLeft:
			gs.handleEventGamePlayerLeft(e.(*event.EventGamePlayerLeft))
		}
	case event.DomainMatch:
		switch e.GetAction() {
		case event.ActionPlayersMatched:
			gs.handleEventPlayersMatched(e.(*event.EventUsersMatched))
		}
	}
}
