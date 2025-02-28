package entity

import (
	"time"

	"github.com/alikarimi999/shahboard/types"
)

type User struct {
	ID        types.ObjectId
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(email string) *User {
	return &User{
		ID:        types.NewObjectId(),
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
