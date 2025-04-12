package ws

import (
	"context"
	"time"

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

			// add userId of sessions that are viewing a game
			// there are multiple approaches to do this, but this approach
			// balances efficiency and scalability and minimizes the need for complex logic.
			viewGamesUsers := make(map[types.ObjectId][]types.ObjectId)
			ss := s.sm.getAll()
			for _, se := range ss {
				gamesId := se.getAllViewGames()
				if len(gamesId) > 0 {
					for _, gameId := range gamesId {
						viewGamesUsers[gameId] = append(viewGamesUsers[gameId], se.userId)
					}
				}
			}

			if err := s.cache.addToGamesViwersList(context.Background(), viewGamesUsers); err != nil {
				s.l.Error(err.Error())
				continue
			}

			// get the updated list of viewers of games from the cache again,
			// to get data that saved in the cache by other instances of wsGateway
			viewersList, err := s.cache.getAllGamesViewersList(context.Background())
			if err != nil {
				s.l.Error(err.Error())
				continue
			}

			for _, se := range ss {
				viewers := viewersList[se.playGameId.Load()]
				if len(viewers) > 0 {
					se.sendViwersList(se.playGameId.Load(), viewers)
				}

				gamesId := se.getAllViewGames()

				if len(gamesId) > 0 {
					for _, gameId := range gamesId {
						viewers := viewersList[gameId]
						if len(viewers) > 0 {
							se.sendViwersList(gameId, viewers)
						}
					}
				}

			}

		case <-cleanCachTicker.C:

			// remove list of viewers of games that ended
			// it's better to be done in only one instance of wsGateway with a master/slave mechanism!
			gamesId := s.em.getAll()
			if len(gamesId) > 0 {
				err := s.cache.removeGamesViewersLists(context.Background(), gamesId)
				if err != nil {
					s.l.Error(err.Error())
				} else {
					s.em.remove(gamesId)
				}
			}

			ss := s.sm.getAll()
			if len(ss) == 0 {
				continue
			}
			// update sessions timestamp in redis cache, every 1 minute
			// this should be done in every wsGateway instances
			if err := s.cache.updateSessionsTimestamp(context.Background(), ss...); err != nil {
				s.l.Error(err.Error())
				continue
			}

			// remove sessions that have not updated their timestamp in the last 3 minutes
			// it's better to be done in only one instance of wsGateway with a master/slave mechanism!
			//
			// this is a simple approach to clean cache from expired sessions
			// that failed to remove after disconnection of client by any reason.
			if err := s.cache.deleteExpiredSessions(context.Background(), sessionCacheTTL); err != nil {
				s.l.Error(err.Error())
				continue
			}

		}

	}

}
