package http

import (
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

func (h *Handler) setupRatingRoutes() {
	r := h.Group("/rating")
	r.GET("/:userId", h.getUserRating)
	r.GET("/history/:userId", h.getUserRatingHistory)
}
func (h *Handler) getUserRating(c *gin.Context) {
	sid := c.Param("userId")

	userId, err := types.ParseObjectId(sid)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	rating, err := h.rating.GetUserRating(c, userId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, UserRatingResponse{
		CurrentScore: rating.CurrentScore,
		BestScore:    rating.BestScore,
		GamesPlayed:  rating.GamesPlayed,
		GamesWon:     rating.GamesWon,
		GamesLost:    rating.GamesLost,
		GamesDraw:    rating.GamesDraw,
		LastUpdated:  rating.LastUpdated.Unix(),
	})
}

func (h *Handler) getUserRatingHistory(c *gin.Context) {
	sid := c.Param("userId")
	userId, err := types.ParseObjectId(sid)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	history, err := h.rating.GetUserChangeHistory(c, userId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	res := UserRatingHistoryResponse{}

	for _, change := range history {
		res.History = append(res.History, UserGameEloChange{
			UserId:     change.UserId.String(),
			GameId:     change.GameId.String(),
			OpponentId: change.OpponentId.String(),
			Change:     change.EloChange,
			Timestamp:  change.UpdatedAt.Unix(),
		})
	}

	c.JSON(200, res)
}
