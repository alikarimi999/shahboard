package types

type User struct {
	ID    ObjectId `json:"id"`
	Email string   `json:"email"`
	Score int64    `json:"score"`
}
