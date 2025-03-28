package http

type UserRatingResponse struct {
	CurrentScore int64 `json:"current_score"`
	BestScore    int64 `json:"best_score"`
	GamesPlayed  int64 `json:"games_played"`
	GamesWon     int64 `json:"games_won"`
	GamesLost    int64 `json:"games_lost"`
	GamesDraw    int64 `json:"games_draw"`
	LastUpdated  int64 `json:"last_updated"`
}

type UserInfoResponse struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	AvatarUrl    string `json:"avatar_url"`
	Bio          string `json:"bio"`
	Country      string `json:"country"`
	CreatedAt    int64  `json:"created_at"`
	LastActiveAt int64  `json:"last_active_at"`
}

type UserRatingHistoryResponse struct {
	History []UserGameEloChange `json:"history"`
}

type UserGameEloChange struct {
	UserId     string `json:"user_id"`
	GameId     string `json:"game_id"`
	OpponentId string `json:"opponent_id"`
	Change     int64  `json:"change"`
	Timestamp  int64  `json:"timestamp"`
}
