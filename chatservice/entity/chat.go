package entity

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/types"
)

type ChatStatus uint8

const (
	ChatStatusActive ChatStatus = iota + 1
	ChatStatusDeactive
)

type Message struct {
	ID        int            `json:"id"`
	SenderId  types.ObjectId `json:"sender_id"`
	Content   string         `json:"content"`
	Timestamp time.Time      `json:"timestamp"`
}

type Messages map[int]Message

type Chat struct {
	id        types.ObjectId
	status    ChatStatus
	player1   types.Player
	player2   types.Player
	mu        sync.RWMutex
	messages  Messages
	createdAt time.Time
	updatedAt time.Time
}

func NewChat(id types.ObjectId, player1, player2 types.Player) *Chat {
	if id.IsZero() {
		id = types.NewObjectId()
	}

	return &Chat{
		id:        id,
		status:    ChatStatusActive,
		player1:   player1,
		player2:   player2,
		messages:  make(Messages),
		createdAt: time.Now(),
	}
}

func (c *Chat) GetId() types.ObjectId {
	return c.id
}

func (c *Chat) Player1() types.Player {
	return c.player1
}

func (c *Chat) Player2() types.Player {
	return c.player2
}

func (c *Chat) IsOwner(id types.ObjectId) bool {
	return c.player1.ID == id || c.player2.ID == id
}

func (c *Chat) AddMessage(msg *Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	index := len(c.messages) + 1
	msg.ID = index
	c.messages[index] = *msg
	c.updatedAt = time.Now()

}

func (c *Chat) GetMessages() []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	messages := make([]Message, 0, len(c.messages))
	for _, message := range c.messages {
		messages = append(messages, message)
	}

	return messages
}

func (c *Chat) GetStatus() ChatStatus {
	return c.status
}

func (c *Chat) Encode() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	b, _ := json.Marshal(struct {
		Id        types.ObjectId `json:"id"`
		Player1   types.Player   `json:"player1"`
		Player2   types.Player   `json:"player2"`
		Messages  Messages       `json:"messages"`
		CreatedAt time.Time      `json:"createdAt"`
		UpdatedAt time.Time      `json:"updatedAt"`
	}{
		Id:        c.id,
		Player1:   c.player1,
		Player2:   c.player2,
		Messages:  c.messages,
		CreatedAt: c.createdAt,
		UpdatedAt: c.updatedAt,
	})
	return b
}

func DecodeChat(b []byte) (*Chat, error) {
	var data struct {
		Id        types.ObjectId `json:"id"`
		Plyer1    types.Player   `json:"player1"`
		Player2   types.Player   `json:"player2"`
		Messages  Messages       `json:"messages"`
		CreatedAt time.Time      `json:"created_at"`
		UpdatedAt time.Time      `json:"updated_at"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	chat := &Chat{
		id:        data.Id,
		player1:   data.Plyer1,
		player2:   data.Player2,
		messages:  data.Messages,
		createdAt: data.CreatedAt,
		updatedAt: data.UpdatedAt,
	}

	return chat, nil
}
