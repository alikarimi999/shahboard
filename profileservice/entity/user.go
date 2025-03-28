package entity

import (
	"time"

	"github.com/alikarimi999/shahboard/types"
)

type UserInfo struct {
	ID           types.ObjectId
	Email        string
	Name         string
	AvatarUrl    string
	Bio          string
	Country      string
	CreatedAt    time.Time
	LastActiveAt time.Time
}
