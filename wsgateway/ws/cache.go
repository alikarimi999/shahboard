package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
	"github.com/redis/go-redis/v9"
)

// Lua scripts for atomic operations with hashes
const (
	addSessionHashScript = `
local key = KEYS[1]
local field = ARGV[1]
local value = ARGV[2]
local cap = tonumber(ARGV[3])

local current_size = redis.call('HLEN', key)
if current_size >= cap then
    return 0
end

redis.call('HSET', key, field, value)
return 1
`
)

type redisCache struct {
	c                            *redis.Client
	userSessionsPrefixKey        string
	userGameSessionsHeartbeatKey string
	gameViewersListKey           string
	expirationTime               time.Duration
	userSessionsCap              int

	l log.Logger
}

// newRedisCache initializes a new redisCache instance
func newRedisCache(c *redis.Client, userSessionsCap int, l log.Logger) *redisCache {
	return &redisCache{
		c:                            c,
		userSessionsPrefixKey:        "user_sessions",
		userGameSessionsHeartbeatKey: "user_game_sessions_heartbeat",
		gameViewersListKey:           "game_viewers",
		expirationTime:               time.Hour * 24,
		userSessionsCap:              userSessionsCap,
		l:                            l,
	}
}

// addUserSessionId adds a session to the user's hash if below the session cap
func (c *redisCache) addUserSessionId(ctx context.Context, userId, sessionId types.ObjectId) (bool, error) {

	key := fmt.Sprintf("%s:%s", c.userSessionsPrefixKey, userId)

	sic := &sessionInCache{
		SessionId: sessionId.String(),
		UserId:    userId.String(),
		GameId:    types.ObjectZero.String(),
		UpdatedAt: time.Now(),
	}

	value, err := json.Marshal(sic)
	if err != nil {
		return false, fmt.Errorf("failed to serialize session: %v", err)
	}

	script := redis.NewScript(addSessionHashScript)

	result, err := script.Run(ctx, c.c, []string{key}, sessionId.String(), value, c.userSessionsCap).Int64()
	if err != nil {
		return false, fmt.Errorf("failed to execute add session script: %v", err)
	}

	return result == 1, nil
}

func (c *redisCache) updateUserGameSession(ctx context.Context, s *session) error {
	key0 := fmt.Sprintf("%s:%s", c.userSessionsPrefixKey, s.userId)

	pipe := c.c.Pipeline()

	t := time.Now()
	sic := &sessionInCache{
		SessionId: s.id.String(),
		UserId:    s.userId.String(),
		GameId:    s.playGameId.String(),
		UpdatedAt: t,
	}

	value, err := json.Marshal(sic)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %v", err)
	}

	pipe.HSet(ctx, key0, s.id.String(), value)

	// update heartbeat key if the session is playing a game
	if !s.playGameId.IsZero() {
		key1 := fmt.Sprintf("%s:%s:%s", c.userGameSessionsHeartbeatKey, s.userId, s.playGameId)
		pipe.Set(ctx, key1, t.Unix(), 0)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for updating session: %v", err)
	}

	return nil
}

// updateSessionsTimestamp is needed to update the timestamp of the live sessions
func (c *redisCache) updateSessionsTimestamp(ctx context.Context, ss ...*session) error {
	if len(ss) == 0 {
		return nil
	}

	pipe := c.c.Pipeline()

	t := time.Now()
	for _, s := range ss {
		sic := &sessionInCache{
			SessionId: s.id.String(),
			UserId:    s.userId.String(),
			GameId:    s.playGameId.String(),
			UpdatedAt: t,
		}

		value, err := json.Marshal(sic)
		if err != nil {
			continue
		}

		key := fmt.Sprintf("%s:%s", c.userSessionsPrefixKey, s.userId.String())
		pipe.HSet(ctx, key, s.id.String(), value)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for updating sessions: %v", err)
	}

	return nil
}

func (c *redisCache) updateUserGameSessionsHeartbeat(ctx context.Context, gamesByUserId map[types.ObjectId]types.ObjectId) error {
	if len(gamesByUserId) == 0 {
		return nil
	}

	t := time.Now().Unix()
	pipe := c.c.Pipeline()
	for userId, gameId := range gamesByUserId {
		key := fmt.Sprintf("%s:%s:%s", c.userGameSessionsHeartbeatKey, userId.String(), gameId.String())
		pipe.Set(ctx, key, t, 0)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for updating user game sessions heartbeat: %v", err)
	}

	return nil
}

func (c *redisCache) deleteExpiredUserGameSessionsHeartbeat(ctx context.Context, ttl time.Duration) (map[types.ObjectId]types.ObjectId, error) {
	keys, err := c.c.Keys(ctx, fmt.Sprintf("%s:*", c.userGameSessionsHeartbeatKey)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}

	pipe := c.c.Pipeline()
	expirationThreshold := time.Now().Add(-ttl)
	deletedSessions := make(map[types.ObjectId]types.ObjectId)
	for _, key := range keys {
		t, err := c.c.Get(ctx, key).Int64()
		if err != nil {
			c.l.Error(fmt.Sprintf("failed to get key: %v", err))
			continue
		}

		if t < expirationThreshold.Unix() {
			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				continue
			}

			deletedSessions[types.ObjectId(parts[1])] = types.ObjectId(parts[2])
			pipe.Del(ctx, key)
		}
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Redis pipeline for deleting expired keys: %v", err)
	}
	return deletedSessions, nil
}

// deleteUsersSessions removes multiple sessions from their respective user hashes
// replace lua script with pipeline execution, because multiple key operations with lua script
// is not supported in redis cluster
func (c *redisCache) deleteUsersSessions(ctx context.Context, ss ...*session) error {
	if len(ss) == 0 {
		return nil
	}

	pipe := c.c.Pipeline()

	for _, s := range ss {
		key := fmt.Sprintf("%s:%s", c.userSessionsPrefixKey, s.userId.String())
		pipe.HDel(ctx, key, s.id.String())
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for removing sessions: %v", err)
	}

	return nil
}

func (c *redisCache) deleteExpiredSessions(ctx context.Context, ttl time.Duration) error {
	keys, err := c.c.Keys(ctx, fmt.Sprintf("%s:*", c.userSessionsPrefixKey)).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %v", err)
	}

	pipe := c.c.Pipeline()
	expirationThreshold := time.Now().Add(-ttl)

	for _, key := range keys {
		sessions, err := c.c.HGetAll(ctx, key).Result()
		if err != nil {
			c.l.Error(fmt.Sprintf("failed to get sessions for key %s: %v", key, err))
			continue
		}

		for sessionId, jsonStr := range sessions {
			var sic sessionInCache
			if err := json.Unmarshal([]byte(jsonStr), &sic); err != nil {
				c.l.Debug(fmt.Sprintf("failed to unmarshal session for key %s: %v", key, err))
				continue
			}

			if sic.UpdatedAt.Before(expirationThreshold) {
				pipe.HDel(ctx, key, sessionId)
			}
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for deleting expired sessions: %v", err)
	}

	return nil
}

// countUsersGameSessions returns the number of game sessions for multiple users
// replace lua script with pipeline execution, same as RemoveUsersSessions
func (c *redisCache) countUsersGameSessions(ctx context.Context, userIds ...types.ObjectId) (map[types.ObjectId]int64, error) {
	if len(userIds) == 0 {
		return nil, nil
	}

	pipe := c.c.Pipeline()
	cmds := make(map[types.ObjectId]*redis.MapStringStringCmd) // map by userId

	for _, userId := range userIds {
		key := fmt.Sprintf("%s:%s", c.userSessionsPrefixKey, userId.String())
		cmds[userId] = pipe.HGetAll(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Redis pipeline: %v", err)
	}

	res := make(map[types.ObjectId]int64)
	for userId, cmd := range cmds {
		sessions, err := cmd.Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get session data: %v", err)
		}

		for _, jsonStr := range sessions {
			gameId := extractGameID(jsonStr)
			if gameId != types.ObjectZero.Int64() {
				res[userId]++
			}
		}
	}

	return res, nil
}

func (c *redisCache) addToGameViwersList(ctx context.Context, userId types.ObjectId, gameId ...types.ObjectId) error {
	pipe := c.c.Pipeline()
	for _, id := range gameId {
		pipe.SAdd(ctx, fmt.Sprintf("%s:%s", c.gameViewersListKey, id.String()), userId.String())
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for adding to game viewers list: %v", err)
	}

	return nil
}

func (c *redisCache) addToGamesViwersList(ctx context.Context, gamesUsers map[types.ObjectId][]types.ObjectId) error {
	pipe := c.c.Pipeline()
	for gameId, usersId := range gamesUsers {
		users := make([]string, 0, len(usersId))
		for _, userId := range usersId {
			users = append(users, userId.String())
		}

		pipe.SAdd(ctx, fmt.Sprintf("%s:%s", c.gameViewersListKey, gameId.String()), users)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for adding to game viewers list: %v", err)
	}

	return nil
}

func (c *redisCache) removeFromGameViewersList(ctx context.Context, userId types.ObjectId,
	gameId ...types.ObjectId) error {

	pipe := c.c.Pipeline()
	for _, id := range gameId {
		pipe.SRem(ctx, fmt.Sprintf("%s:%s", c.gameViewersListKey, id.String()), userId.String())
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for removing from game viewers list: %v", err)
	}

	return nil
}

func (c *redisCache) removeGamesViewersLists(ctx context.Context, gamesId []types.ObjectId) error {
	keys := make([]string, 0, len(gamesId))
	for _, id := range gamesId {
		keys = append(keys, fmt.Sprintf("%s:%s", c.gameViewersListKey, id.String()))
	}

	return c.c.Del(ctx, keys...).Err()
}

// func (c *redisCache) getGameViewers(ctx context.Context, gameId types.ObjectId) ([]string, error) {
// 	return c.c.SMembers(ctx, fmt.Sprintf("%s:%s", c.gameViewersListKey, gameId.String())).Result()
// }

func (c *redisCache) getAllGamesViewersList(ctx context.Context) (map[types.ObjectId][]types.ObjectId, error) {
	keys, err := c.c.Keys(ctx, fmt.Sprintf("%s:*", c.gameViewersListKey)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}

	res := make(map[types.ObjectId][]types.ObjectId)
	for _, key := range keys {
		id := strings.TrimPrefix(key, c.gameViewersListKey+":")
		users, err := c.c.SMembers(ctx, key).Result()
		if err != nil {
			c.l.Error(fmt.Sprintf("failed to get users for key %s: %v", key, err))
			continue
		}

		gameId, err := types.ParseObjectId(id)
		if err != nil {
			continue
		}

		usersId := make([]types.ObjectId, 0, len(users))
		for _, user := range users {
			userId, err := types.ParseObjectId(user)
			if err == nil {
				usersId = append(usersId, userId)
			}
		}

		res[gameId] = usersId
	}

	return res, nil
}

func extractGameID(jsonStr string) int64 {
	var sic sessionInCache
	if err := json.Unmarshal([]byte(jsonStr), &sic); err != nil {
		return types.ObjectZero.Int64()
	}

	i, err := strconv.Atoi(sic.GameId)
	if err != nil {
		return types.ObjectZero.Int64()
	}
	return int64(i)
}

type sessionInCache struct {
	SessionId string    `json:"sessionId"`
	UserId    string    `json:"userId"`
	GameId    string    `json:"gameId"`
	UpdatedAt time.Time `json:"updatedAt"`
}
