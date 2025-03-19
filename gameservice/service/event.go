package game

import (
	"context"
	"fmt"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/gameservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

// handleEventUsersMatched handles the event of two players being matched
// and creates a new game if both players do not have any active games.
// This function is idempotent and safe to call concurrently from multiple
// instances of the GameService.
func (s *Service) handleEventUsersMatched(d *event.EventUsersMatchCreated) {
	s.l.Debug(fmt.Sprintf("handling event users matched: '%s' and '%s'", d.User1.ID, d.User2.ID))
	// check if player is already in a game
	if s.checkByPlayer(d.User1.ID) || s.checkByPlayer(d.User2.ID) {
		s.l.Debug("player is already in a game")
		return
	}

	// create a new game
	game := entity.NewGame(d.User1.ID, d.User2.ID, s.cfg.DefaultGameSettings)

	// add the game to the cache
	if ok, err := s.cache.addGame(context.Background(), game); err != nil {
		s.l.Error(err.Error())
		return
	} else if !ok {
		s.l.Debug(fmt.Sprintf("game already exists in cache: '%s'", game.ID()))
		return
	}
	s.l.Debug(fmt.Sprintf("added game to cache: '%s'", game.ID()))

	if !s.addGame(game) {
		s.l.Error(fmt.Sprintf("game '%s' created before", game.ID()))
		return
	}

	// publish the game created event
	if err := s.pub.Publish(event.EventGameCreated{
		GameID:    game.ID(),
		MatchID:   d.ID,
		Player1:   game.Player1(),
		Player2:   game.Player2(),
		Timestamp: time.Now().Unix(),
	}); err != nil {
		s.l.Error(err.Error())
	}
	s.l.Debug(fmt.Sprintf("published game created event: '%s'", game.ID()))

	s.l.Info(fmt.Sprintf("game '%s' created by player '%s' as '%s' and player '%s' as '%s'",
		game.ID(), game.Player1().ID, game.Player1().Color, game.Player2().ID, game.Player2().Color))

	// TODO: add to repository concurrency control
}

func (s *Service) handleEventGamePlayerMoved(d *event.EventGamePlayerMoved) {
	game := s.getGame(d.GameID)
	// if game is not manging by this instance, do nothing
	if game == nil || game.Status() == entity.GameStatusDeactive {
		return
	}

	game.Lock()
	if game.Turn().ID != d.PlayerID {
		game.Unlock()
		s.l.Debug(fmt.Sprintf("it's not player '%s' turn", d.PlayerID))
		return
	}

	if err := game.Move(d.Move); err != nil {
		game.Unlock()
		s.l.Debug(fmt.Sprintf("player '%s' made an invalid move '%s' on game '%s'", d.PlayerID, d.Move, d.GameID))
		return
	}
	game.Unlock()

	if game.Outcome() != entity.NoOutcome {
		if !game.EndGame() {
			return
		}

		if err := s.cache.updateAndDeactivateGame(context.Background(), game); err != nil {
			s.l.Error(err.Error())
			return
		}

		if err := s.pub.Publish(event.EventGameMoveApproved{
			ID:        types.NewObjectId(),
			PlayerID:  d.PlayerID,
			GameID:    d.GameID,
			Move:      d.Move,
			Timestamp: time.Now().Unix(),
		},
			event.EventGameEnded{
				ID:        types.NewObjectId(),
				GameID:    game.ID(),
				Player1:   game.Player1(),
				Player2:   game.Player2(),
				Outcome:   string(game.Outcome()),
				Timestamp: time.Now().Unix(),
			}); err != nil {
			s.l.Error(err.Error())
			return
		}

		s.l.Debug(fmt.Sprintf("published game move approved event: '%s'", game.ID()))
		s.l.Debug(fmt.Sprintf("published game ended event: '%s'", game.ID()))

		s.removeGame(game.ID())

	} else {
		if err := s.cache.updateGameMove(context.Background(), game); err != nil {
			s.l.Debug(err.Error())
			return
		}

		s.l.Debug(fmt.Sprintf("player '%s' made move '%s' on game '%s'", d.PlayerID, d.Move, game.ID()))

		s.pub.Publish(event.EventGameMoveApproved{
			ID:        types.NewObjectId(),
			PlayerID:  d.PlayerID,
			GameID:    d.GameID,
			Move:      d.Move,
			Timestamp: time.Now().Unix(),
		})
		s.l.Debug(fmt.Sprintf("published game move approved event: '%s'", game.ID()))
	}

	// TODO: think about how to update the database
}

func (s *Service) handleEventGamePlayerLeft(d *event.EventGamePlayerLeft) {
	game := s.getGame(d.GameID)
	if game == nil || game.Status() == entity.GameStatusDeactive {
		return
	}

	if !game.PlayerLeft(d.PlayerID) {
		return
	}

	s.l.Debug(fmt.Sprintf("player '%s' left game '%s'", d.PlayerID, d.GameID))

	if err := s.cache.updateAndDeactivateGame(context.Background(), game); err != nil {
		s.l.Error(err.Error())
		return
	}

	s.removeGame(d.GameID)
	s.l.Debug(fmt.Sprintf("removed game: '%s'", d.GameID))

	s.pub.Publish(event.EventGameEnded{
		ID:        types.NewObjectId(),
		GameID:    d.GameID,
		Player1:   game.Player1(),
		Player2:   game.Player2(),
		Outcome:   game.Outcome().String(),
		Desc:      entity.EndDescriptionPlayerLeft.String(),
		Timestamp: time.Now().Unix(),
	})
	s.l.Debug(fmt.Sprintf("published game ended event: '%s'", d.GameID))

}

func (gs *Service) handleEvents(e event.Event) {
	switch e.GetTopic().Domain() {
	case event.DomainGame:
		res := e.GetTopic().Resource()
		gameId, err := types.ParseObjectId(res)
		if err != nil {
			return
		}

		if !gs.gameExists(gameId) {
			return
		}

		switch e.GetTopic().Action() {
		case event.ActionGamePlayerMoved:
			gs.handleEventGamePlayerMoved(e.(*event.EventGamePlayerMoved))
		case event.ActionGamePlayerLeft:
			gs.handleEventGamePlayerLeft(e.(*event.EventGamePlayerLeft))
		}

	case event.DomainMatch:
		switch e.GetTopic().Action() {
		case event.ActionCreated:
			gs.handleEventUsersMatched(e.(*event.EventUsersMatchCreated))
		}
	}
}
