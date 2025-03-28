package http

import (
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

func (h *Handler) setupUserRoutes() {
	u := h.Group("/user")
	u.GET("/:userId", h.getUserInfo)
}

func (h *Handler) getUserInfo(c *gin.Context) {
	sid := c.Param("userId")
	userId, err := types.ParseObjectId(sid)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.user.GetUserInfo(c, userId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	c.JSON(200, UserInfoResponse{
		Name:         user.Name,
		Email:        user.Email,
		AvatarUrl:    user.AvatarUrl,
		Bio:          user.Bio,
		Country:      user.Country,
		CreatedAt:    user.CreatedAt.Unix(),
		LastActiveAt: user.LastActiveAt.Unix(),
	})
}
