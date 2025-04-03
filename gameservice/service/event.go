package game

import (
	"context"
	"fmt"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/gameservice/entity"
	"github.com/alikarimi999/shahboard/types"
	"github.com/notnil/chess"
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
	game := entity.NewGame(d.User1, d.User2, s.cfg.DefaultGameSettings)

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

	if err := game.Move(d.PlayerID, d.Move, d.Index); err != nil {
		s.l.Debug(fmt.Sprintf("player '%s' made an invalid move '%s' on game '%s'", d.PlayerID, d.Move, d.GameID))
		return
	}

	if game.Outcome() != types.NoOutcome {
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
			Index:     d.Index,
			Timestamp: time.Now().Unix(),
		},
			event.EventGameEnded{
				ID:        types.NewObjectId(),
				GameID:    game.ID(),
				Player1:   game.Player1(),
				Player2:   game.Player2(),
				Outcome:   game.Outcome(),
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
			Index:     d.Index,
			Timestamp: time.Now().Unix(),
		})
		s.l.Debug(fmt.Sprintf("published game move approved event: '%s'", game.ID()))
	}

	// TODO: think about how to update the database
}

func (s *Service) handleEventGamePlayerClaimDraw(d *event.EventGamePlayerClaimDraw) {
	game := s.getGame(d.GameID)
	if game == nil || game.Status() == entity.GameStatusDeactive {
		return
	}

	t := time.Now().Unix()
	ea := event.EventGamePlayerClaimDrawApproved{
		ID:        types.NewObjectId(),
		ClaimID:   d.ID,
		GameID:    d.GameID,
		PlayerID:  d.PlayerID,
		Method:    d.Method,
		Timestamp: t,
	}
	switch d.Method {
	case chess.DrawOffer:
		// if offer draw is approved, the service should send the approve event to the other player
		// and wait for the other player to accept the offer
		if !game.OfferDraw(d.PlayerID) {
			return
		}

		if err := s.pub.Publish(ea); err != nil {
			s.l.Error(err.Error())
			return
		}
	case chess.ThreefoldRepetition, chess.FiftyMoveRule:

		// if claim draw for these methods is approved, the service should send the approve event
		// and end the game with the outcome of the draw
		if !game.ClaimDraw(d.PlayerID, d.Method) {
			return
		}

		if err := s.cache.updateAndDeactivateGame(context.Background(), game); err != nil {
			s.l.Error(err.Error())
			return
		}

		if err := s.pub.Publish(ea, event.EventGameEnded{
			ID:        types.NewObjectId(),
			GameID:    game.ID(),
			Player1:   game.Player1(),
			Player2:   game.Player2(),
			Outcome:   game.Outcome(),
			Timestamp: t,
		}); err != nil {
			s.l.Error(err.Error())
			return
		}

		s.removeGame(game.ID())
	default:
		return
	}

}

func (s *Service) handleEventGamePlayerResponsedDrawOffer(d *event.EventGamePlayerResponsedDrawOffer) {
	game := s.getGame(d.GameID)
	if game == nil || game.Status() == entity.GameStatusDeactive {
		return
	}

	if !d.Accept {
		game.RejectDraw(d.PlayerID)
		return
	}

	if !game.AcceptDraw(d.PlayerID) {
		return
	}

	if err := s.cache.updateAndDeactivateGame(context.Background(), game); err != nil {
		s.l.Error(err.Error())
		return
	}

	if err := s.pub.Publish(event.EventGameEnded{
		ID:        types.NewObjectId(),
		GameID:    game.ID(),
		Player1:   game.Player1(),
		Player2:   game.Player2(),
		Outcome:   game.Outcome(),
		Timestamp: time.Now().Unix(),
	}); err != nil {
		s.l.Error(err.Error())
		return
	}

	s.removeGame(game.ID())
}
func (s *Service) handleEventGamePlayerResigned(d *event.EventGamePlayerResigned) {
	game := s.getGame(d.GameID)
	if game == nil || game.Status() == entity.GameStatusDeactive {
		return
	}

	if !game.Resign(d.PlayerID) {
		return
	}

	if err := s.cache.updateAndDeactivateGame(context.Background(), game); err != nil {
		s.l.Error(err.Error())
		return
	}

	s.removeGame(d.GameID)

	if err := s.pub.Publish(event.EventGameEnded{
		ID:        types.NewObjectId(),
		GameID:    d.GameID,
		Player1:   game.Player1(),
		Player2:   game.Player2(),
		Outcome:   game.Outcome(),
		Desc:      entity.EndDescriptionPlayerResigned.String(),
		Timestamp: time.Now().Unix(),
	}); err != nil {
		s.l.Error(fmt.Sprintf("failed to publish game ended event: '%s'", d.GameID))
	}

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

	if err := s.pub.Publish(event.EventGameEnded{
		ID:        types.NewObjectId(),
		GameID:    d.GameID,
		Player1:   game.Player1(),
		Player2:   game.Player2(),
		Outcome:   game.Outcome(),
		Desc:      entity.EndDescriptionPlayerLeft.String(),
		Timestamp: time.Now().Unix(),
	}); err != nil {
		s.l.Error(fmt.Sprintf("failed to publish game ended event: '%s'", d.GameID))
	}
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
		case event.ActionGamePlayerClaimDraw:
			gs.handleEventGamePlayerClaimDraw(e.(*event.EventGamePlayerClaimDraw))
		case event.ActionGamePlayerResponsedDrawOffer:
			gs.handleEventGamePlayerResponsedDrawOffer(e.(*event.EventGamePlayerResponsedDrawOffer))
		case event.ActionGamePlayerResigned:
			gs.handleEventGamePlayerResigned(e.(*event.EventGamePlayerResigned))
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
