package http

func (r *Router) setupUserRoutes() {

	// r.gin.GET("pgn/:id", r.getGamePGN)
	r.gin.GET("/live/", r.getLiveGames)
	r.gin.GET("/live/data", r.getLiveGamesData)
	r.gin.GET("/live/user/:id", r.getLiveGameByUserId)

	// The service isn't statless and this endpoint can't be used in a multi instance setup.
	r.gin.GET("/live/resign/:gameId", r.resignByPlayer)
	// r.gin.POST("/fen", r.getGamesFen)

}
