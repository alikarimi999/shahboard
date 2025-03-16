package ws

import (
	"context"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
)

// The manageSessionsState function effectively manages WebSocket sessions by:
// - Disconnecting inactive clients every 5 seconds based on heartbeats.
// - Updating session timestamps and cleaning expired sessions every minute.
// - Tracking game sessions and notifying the game service when players leave.
//
// This function needs some improvements and maybe separate each task into its own function
// for better modularity and maintainability.
func (s *Server) manageSessionsState() {
	sessionCacheTTL := time.Minute * 3
	gameSessionTTL := time.Minute * 2
	cleanCachTicker := time.NewTicker(time.Minute * 1)
	pingTicker := time.NewTicker(time.Second * 5)
	pingIntervalDisconnectedSession := time.Second * 10

	for {
		select {
		case t := <-pingTicker.C:
			// remove sessions that have not sent a heartbeat in the last 10 seconds
			disconnectedSessions := []*session{}
			expireTreshold := t.Add(-pingIntervalDisconnectedSession)
			for _, se := range s.sm.getAll() {
				lh := se.lastHeartBeat.Load()
				if lh.Before(expireTreshold) {
					disconnectedSessions = append(disconnectedSessions, se)
					continue
				}
			}

			if len(disconnectedSessions) > 0 {
				s.stopSessions(disconnectedSessions...)
			}

		case <-cleanCachTicker.C:
			ss := s.sm.getAll()
			// update sessions timestamp in redis cache, every 1 minute
			// this should be done in every wsGateway instances
			if err := s.cache.updateSessionsTimestamp(context.Background(), ss...); err != nil {
				s.l.Error(err.Error())
				continue
			}

			// remove sessions that have not updated their timestamp in the last 3 minutes
			// it's better to be done in only one instance of wsGateway with a master/slave mechanism!
			if err := s.cache.deleteExpiredSessions(context.Background(), sessionCacheTTL); err != nil {
				s.l.Error(err.Error())
				continue
			}

			liveGameSessions := make(map[types.ObjectId]types.ObjectId)
			for _, se := range ss {
				if se.playGameId != types.ObjectZero {
					liveGameSessions[se.userId] = se.playGameId
				}
			}

			// update game_sessions_heartbeat cache, every 1 minute
			// this should be done in every wsGateway instances
			// this is used to detect games that have been left by one of the players for a while
			if len(liveGameSessions) > 0 {
				if err := s.cache.updateUserGameSessionsHeartbeat(context.Background(), liveGameSessions); err != nil {
					s.l.Error(err.Error())
					continue
				}
			}

			// remove expired game_sessions_heartbeat cache, that have not been updated in the last 2 minutes
			// it's better to be done in only one instance of wsGateway with a master/slave mechanism!
			deletedSessions, err := s.cache.deleteExpiredUserGameSessionsHeartbeat(context.Background(), gameSessionTTL)
			if err != nil {
				s.l.Error(err.Error())
				continue
			}

			// publish event to game service to notify that the game has been left by one of the players
			events := make([]event.Event, 0, len(deletedSessions))
			for userId, gameId := range deletedSessions {
				events = append(events, event.EventGamePlayerLeft{
					GameID:    gameId,
					PlayerID:  userId,
					Timestamp: time.Now().Unix(),
				})
			}

			if len(events) > 0 {
				if err := s.p.Publish(events...); err != nil {
					s.l.Error(err.Error())
				}
			}

		}

	}

}
