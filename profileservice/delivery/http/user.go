package http

import (
	"github.com/alikarimi999/shahboard/pkg/elo"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

func (h *Handler) setupUserRoutes() {
	u := h.Group("/users")
	u.GET("/:userId", h.getUserInfo)
}

func (h *Handler) getUserInfo(c *gin.Context) {
	sid := c.Param("userId")
	userId, err := types.ParseObjectId(sid)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	u, r, err := h.user.GetUserInfo(c, userId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if u == nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	var score int64
	var level string
	if r != nil {
		score = r.CurrentScore
		level = elo.GetPlayerLevel(score).String()
	}

	c.JSON(200, UserInfoResponse{
		Name:         u.Name,
		Email:        u.Email,
		AvatarUrl:    u.AvatarUrl,
		Bio:          u.Bio,
		Country:      u.Country,
		Score:        score,
		Level:        level,
		CreatedAt:    u.CreatedAt.Unix(),
		LastActiveAt: u.LastActiveAt.Unix(),
	})
}
