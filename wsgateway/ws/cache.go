package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
	"github.com/redis/go-redis/v9"
)

// Lua scripts for atomic operations with hashes
const (
	addUserSessionScript = `
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

	removeGameIdScript = `
		local key = KEYS[1]
		local sessionId = ARGV[1]
		local zeroGameIdJson = ARGV[2]

		redis.call("HSET", key, sessionId, zeroGameIdJson)

		local sessions = redis.call("HVALS", key)
		local count = 0

		for _, v in ipairs(sessions) do
			local ok, parsed = pcall(cjson.decode, v)
			if ok and parsed["gameId"] and parsed["gameId"] ~= "" and parsed["gameId"] ~= "" then
				count = count + 1
			end
		end

		return count
	`
)

type redisCache struct {
	c                  *redis.Client
	userSessions       string
	gameViewersListKey string
	expirationTime     time.Duration
	userSessionsCap    int

	l log.Logger
}

// newRedisCache initializes a new redisCache instance
func newRedisCache(c *redis.Client, userSessionsCap int, l log.Logger) *redisCache {
	return &redisCache{
		c:                  c,
		userSessions:       "user_sessions",
		gameViewersListKey: "game_viewers",
		expirationTime:     time.Hour * 24,
		userSessionsCap:    userSessionsCap,
		l:                  l,
	}
}

// addUserSessionId adds a session to the user's hash if below the session cap
func (c *redisCache) addUserSessionId(ctx context.Context, userId, sessionId types.ObjectId) (bool, error) {

	key := fmt.Sprintf("%s:%s", c.userSessions, userId)

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

	script := redis.NewScript(addUserSessionScript)

	result, err := script.Run(ctx, c.c, []string{key}, sessionId.String(), value, c.userSessionsCap).Int64()
	if err != nil {
		return false, fmt.Errorf("failed to execute add session script: %v", err)
	}

	return result == 1, nil
}

func (c *redisCache) addGameIdToUserSessions(ctx context.Context, userId, sessionId, gameId types.ObjectId) error {
	key := fmt.Sprintf("%s:%s", c.userSessions, userId)
	sic := &sessionInCache{
		SessionId: sessionId.String(),
		UserId:    userId.String(),
		GameId:    gameId.String(),
		UpdatedAt: time.Now(),
	}

	value, err := json.Marshal(sic)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %v", err)
	}

	return c.c.HSet(ctx, key, sessionId.String(), value).Err()
}

func (c *redisCache) removeGameIdFromUserSessions(ctx context.Context, userId, sessionId types.ObjectId) (int64, error) {
	key := fmt.Sprintf("%s:%s", c.userSessions, userId)

	sic := &sessionInCache{
		SessionId: sessionId.String(),
		UserId:    userId.String(),
		GameId:    types.ObjectZero.String(),
		UpdatedAt: time.Now(),
	}
	jsonValue, err := json.Marshal(sic)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize session: %v", err)
	}

	res, err := c.c.Eval(ctx, removeGameIdScript, []string{key}, sessionId.String(), string(jsonValue)).Result()
	if err != nil {
		return 0, fmt.Errorf("lua script failed: %v", err)
	}

	count, ok := res.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type from lua script")
	}

	return count, nil
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
			GameId:    s.playGameId.Load().String(),
			UpdatedAt: t,
		}

		value, err := json.Marshal(sic)
		if err != nil {
			continue
		}

		key := fmt.Sprintf("%s:%s", c.userSessions, s.userId.String())
		pipe.HSet(ctx, key, s.id.String(), value)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for updating sessions: %v", err)
	}

	return nil
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
		key := fmt.Sprintf("%s:%s", c.userSessions, s.userId.String())
		pipe.HDel(ctx, key, s.id.String())
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline for removing sessions: %v", err)
	}

	return nil
}

func (c *redisCache) deleteExpiredSessions(ctx context.Context, ttl time.Duration) error {
	keys, err := c.c.Keys(ctx, fmt.Sprintf("%s:*", c.userSessions)).Result()
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

func (c *redisCache) countAllGamesViewers(ctx context.Context) (map[types.ObjectId]int64, error) {
	keys, err := c.c.Keys(ctx, fmt.Sprintf("%s:*", c.gameViewersListKey)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}

	res := make(map[types.ObjectId]int64)
	for _, key := range keys {
		i, err := c.c.SCard(ctx, key).Result()
		if err != nil {
			c.l.Error(fmt.Sprintf("failed to get the size of '%s': %v", key, err))
			continue
		}

		gameId, err := types.ParseObjectId(strings.TrimPrefix(key, c.gameViewersListKey+":"))
		if err != nil {
			continue
		}

		res[gameId] = i
	}

	return res, nil
}

type sessionInCache struct {
	SessionId string    `json:"sessionId"`
	UserId    string    `json:"userId"`
	GameId    string    `json:"gameId"`
	UpdatedAt time.Time `json:"updatedAt"`
}
