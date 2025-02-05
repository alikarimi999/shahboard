package types

type User struct {
	ID    ObjectId `json:"id"`
	Level Level    `json:"level"`
}
