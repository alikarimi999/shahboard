package main

type Config struct {
	Server        string `json:"server"`
	StockfishPath string `json:"stockfish_path"`
	Bots          []Bot  `json:"bots"`
}

type Bot struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
