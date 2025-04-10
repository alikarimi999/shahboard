package config

type Config struct {
	Server         string `json:"server"`
	StockfishPath  string `json:"stockfish_path"`
	BotsNum        int    `json:"bots_num"`
	Local          bool   `json:"local"`
	AuthService    string `json:"auth_service"`
	MatchService   string `json:"match_service"`
	GameService    string `json:"game_service"`
	ProfileService string `json:"profile_service"`
	WsService      string `json:"ws_service"`
}
