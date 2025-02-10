package service

import (
	"context"
	"fmt"
	"time"

	"github.com/alikarimi999/shahboard/chatservice/entity"
	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
)

func (s *Service) handleEvents(e event.Event) {
	switch e.GetTopic().Domain() {
	case event.DomainGame:
		switch e.GetAction() {
		case event.ActionCreated:
			s.handleGameCreated(e.(*event.EventGameCreated))
		case event.ActionEnded:
			s.handleGameEnded(e.(*event.EventGameEnded))
		}
	case event.DomainGameChat:
		switch e.GetAction() {
		case event.ActionMsgSent:
			s.handleMsgSent(e.(*event.EventGameChatMsgeSent))
		}

	}
}

func (s *Service) handleGameCreated(e *event.EventGameCreated) {
	ok, err := s.CreateGameChat(context.Background(), e.GameID, e.Player1, e.Player2)
	if err != nil {
		s.l.Error(fmt.Sprintf("failed to create game chat, ERR: %s", err.Error()))
		return
	}

	if !ok {
		s.l.Debug("failed to create game chat, ERR: chat already exists")
		return
	}

	if err := s.pub.Publish(event.EventGameChatCreated{
		ID:        types.NewObjectId(),
		GameID:    e.GameID,
		MatchID:   e.MatchID,
		Player1:   e.Player1,
		Player2:   e.Player2,
		Timestamp: e.Timestamp,
	}); err != nil {
		s.l.Error(fmt.Sprintf("failed to publish game chat created event, ERR: %s", err.Error()))
		return
	}

	s.sm.AddSubscription(s.sub.Subscribe(event.TopicGameChat.WithResource(e.GameID.String())))
	s.l.Debug(fmt.Sprintf("game chat created, gameID: %s", e.GameID))
}

func (s *Service) handleMsgSent(e *event.EventGameChatMsgeSent) {
	c := s.cm.getChat(e.GameID)
	if c == nil {
		s.l.Debug(fmt.Sprintf("chat not found, gameID: %s", e.GameID))
		return
	}

	if !c.IsOwner(e.SenderID) {
		s.l.Debug(fmt.Sprintf("invalid sender, gameID: %s, senderID: %s", e.GameID, e.SenderID))
		return
	}

	if e.Content == "" {
		s.l.Debug(fmt.Sprintf("empty message, gameID: %s, senderID: %s", e.GameID, e.SenderID))
		return
	}

	msg := &entity.Message{
		SenderId:  e.SenderID,
		Content:   e.Content,
		Timestamp: time.Now(),
	}

	c.AddMessage(msg)

	if err := s.pub.Publish(event.EventGameChatMsgApproved{
		ID:        types.NewObjectId(),
		GameID:    e.GameID,
		Message:   *msg,
		Timestamp: time.Now().Unix(),
	}); err != nil {
		s.l.Error(fmt.Sprintf("failed to publish game chat message approved event, ERR: %s", err.Error()))
		return
	}

	s.l.Debug(fmt.Sprintf("message sent, gameID: %s, senderID: %s", e.GameID, e.SenderID))
}

func (s *Service) handleGameEnded(e *event.EventGameEnded) {
	c := s.cm.getChat(e.GameID)
	if c == nil {
		return
	}

	s.cm.removeChat(e.GameID)

	if err := s.cache.deleteGameChat(context.Background(), e.GameID); err != nil {
		s.l.Error(fmt.Sprintf("failed to delete game chat, ERR: %s", err.Error()))
	}

	if err := s.pub.Publish(event.EventGameChatEnded{
		ID:        types.NewObjectId(),
		GameID:    e.GameID,
		Player1:   c.Player1(),
		Player2:   c.Player2(),
		Timestamp: e.Timestamp,
	}); err != nil {
		s.l.Error(fmt.Sprintf("failed to publish game chat ended event, ERR: %s", err.Error()))
		return
	}

	s.l.Debug(fmt.Sprintf("game chat removed, gameID: %s", e.GameID))
}
