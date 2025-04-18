package config

type Config struct {
	Server         string `json:"server"`
	StockfishPath  string `json:"stockfish_path"`
	Local          bool   `json:"local"`
	AuthService    string `json:"auth_service"`
	MatchService   string `json:"match_service"`
	GameService    string `json:"game_service"`
	ProfileService string `json:"profile_service"`
	WsService      string `json:"ws_service"`
	StopChance     int    `json:"stop_chance"`
	ResignChance   int    `json:"resign_chance"`
}
