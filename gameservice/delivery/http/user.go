package http

func (r *Router) setupUserRoutes() {
	g := r.gin.Group("/games")
	{
		// g.GET("pgn/:id", r.getGamePGN)
		g.GET("/live/", r.getLiveGames)
		g.GET("/live/data", r.getLiveGamesData)
		g.GET("/live/user/:id", r.getLiveGameByUserId)

		// g.POST("/fen", r.getGamesFen)
	}
}
