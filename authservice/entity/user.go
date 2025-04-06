package entity

import (
	"time"

	"github.com/alikarimi999/shahboard/types"
)

type User struct {
	ID        types.ObjectId
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(email, pass string) *User {
	return &User{
		ID:        types.NewObjectId(),
		Email:     email,
		Password:  pass,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
