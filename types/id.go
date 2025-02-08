package types

import (
	"strconv"
	"time"

	"math/rand"
)

type ObjectId string

func NewObjectId() ObjectId {
	// Get the current timestamp (4 bytes)
	timestamp := uint32(time.Now().Unix())

	// Generate a random 32-bit integer (4 bytes)
	random := uint32(rand.Intn(0xFFFFFFFF))

	// Combine the timestamp and random value into a single int64
	id := (int64(timestamp) << 32) | int64(random)

	return ObjectId(strconv.FormatInt(id, 10))
}

func ZeroObjectId() ObjectId {
	return ObjectId("")
}

func (id ObjectId) IsZero() bool {
	return id == "" || id == "0"
}

func (id ObjectId) Int64() int64 {
	nid, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return 0
	}
	return nid
}

func (id ObjectId) String() string {
	return string(id)
}

func ParseObjectId(s string) (ObjectId, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return "", err
	}
	return ObjectId(strconv.FormatInt(id, 10)), nil
}
