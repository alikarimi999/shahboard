package http

import (
	"github.com/alikarimi999/shahboard/pkg/elo"
	"github.com/alikarimi999/shahboard/pkg/middleware"
	"github.com/alikarimi999/shahboard/profileservice/service/user"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

func (h *Handler) setupUserRoutes() {
	u := h.Group("/users")
	u.GET("/:userId", h.getUserInfo)
	u.PATCH("/", h.updateUserInfo)
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

func (h *Handler) updateUserInfo(c *gin.Context) {
	usr, ok := middleware.ExtractUser(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req user.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.user.UpdateUser(c, usr.ID, req); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User updated successfully"})
}
