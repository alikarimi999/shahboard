package types

import (
	"fmt"
	"time"

	"math/rand"
)

type ObjectId int64

func NewObjectId() ObjectId {
	// Get the current timestamp (4 bytes)
	timestamp := uint32(time.Now().Unix())

	// Generate a random 32-bit integer (4 bytes)
	random := uint32(rand.Intn(0xFFFFFFFF))

	// Combine the timestamp and random value into a single int64
	id := (int64(timestamp) << 32) | int64(random)

	return ObjectId(id)
}

func (id ObjectId) String() string {
	return fmt.Sprintf("%d", id)
}
