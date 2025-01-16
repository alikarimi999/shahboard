package http

import (
	"net/http"

	"github.com/alikarimi999/shahboard/pkg/middleware"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

func (r *Router) setupUserRoutes() {
	u := r.gin.Group("/user", middleware.ParsUserHeader())
	{
		u.GET("/match", r.newMatchRequest)
	}
}

func (r *Router) newMatchRequest(c *gin.Context) {
	u := getUser(c)

	m, err := r.s.NewMatchRequest(c.Request.Context(), u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

func getUser(c *gin.Context) types.User {
	u, _ := c.Get("user")
	return u.(types.User)
}
