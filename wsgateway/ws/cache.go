package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alikarimi999/shahboard/types"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	c                    *redis.Client
	gameSessionPrefixKey string
	expirationTime       time.Duration
}

func newRedisCache(c *redis.Client) *redisCache {
	return &redisCache{
		c:                    c,
		gameSessionPrefixKey: "user_game_session",
		expirationTime:       time.Hour * 24,
	}
}

func (c *redisCache) SaveSessionState(ctx context.Context, sess *session) error {
	s := &sessionInCache{
		SessionId: sess.id,
		UserId:    sess.userId,
		GameId:    sess.gameId,
		Role:      sess.role,
		Closed:    sess.isClosed(),
	}

	return c.c.Set(ctx, fmt.Sprintf("%s:%s", c.gameSessionPrefixKey, s.UserId), s.encode(), c.expirationTime).Err()
}

func (c *redisCache) SaveSessionsState(ctx context.Context, sess ...*session) error {
	ss := make(map[string]interface{})
	for _, s := range sess {
		sic := &sessionInCache{
			SessionId: s.id,
			UserId:    s.userId,
			GameId:    s.gameId,
			Role:      s.role,
			Closed:    s.isClosed(),
		}
		ss[fmt.Sprintf("%s:%s", c.gameSessionPrefixKey, s.userId)] = sic.encode()
	}

	return c.c.MSet(ctx, ss).Err()
}

func (c *redisCache) DeleteSessions(ctx context.Context, usersId ...types.ObjectId) error {
	keys := make([]string, len(usersId))
	for i, userId := range usersId {
		keys[i] = fmt.Sprintf("%s:%s", c.gameSessionPrefixKey, userId)
	}
	return c.c.Del(ctx, keys...).Err()
}

func (c *redisCache) GetSession(ctx context.Context, userId types.ObjectId) (*sessionInCache, error) {
	session, err := c.c.Get(ctx, fmt.Sprintf("%s:%s", c.gameSessionPrefixKey, userId)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}
	sess := &sessionInCache{}
	err = sess.decode([]byte(session))
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// func (c *redisCache) DeleteSession(ctx context.Context, userId types.ObjectId) error {
// 	return c.c.Del(ctx, fmt.Sprintf("%s:%s", c.gameSessionPrefixKey, userId)).Err()
// }

// func (c *redisCache) DeleteSessions(ctx context.Context, userIds ...types.ObjectId) error {
// 	keys := make([]string, len(userIds))
// 	for i, userId := range userIds {
// 		keys[i] = fmt.Sprintf("%s:%s", c.gameSessionPrefixKey, userId)
// 	}
// 	return c.c.Del(ctx, keys...).Err()
// }

type sessionInCache struct {
	SessionId types.ObjectId
	UserId    types.ObjectId
	GameId    types.ObjectId
	Role      gameRole
	Closed    bool
}

func (s *sessionInCache) encode() []byte {
	b, _ := json.Marshal(s)
	return b
}

func (s *sessionInCache) decode(b []byte) error {
	return json.Unmarshal(b, s)
}
