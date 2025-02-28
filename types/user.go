package types

type User struct {
	ID    ObjectId `json:"id"`
	Email string   `json:"email"`
	Level Level    `json:"level"`
}
