package types

type User struct {
	ID      ObjectId `json:"id"`
	Email   string   `json:"email"`
	IsGuest bool     `json:"is_guest"`
	Score   int64    `json:"score"`
}
