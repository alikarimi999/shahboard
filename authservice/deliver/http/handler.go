package http

import (
	"github.com/alikarimi999/shahboard/authservice/service"
	"github.com/alikarimi999/shahboard/pkg/router"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	*router.Router
	s *service.AuthService
}

func NewHandler(cfg router.Config, s *service.AuthService) (*Handler, error) {
	r, err := router.NewRouter(cfg)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		Router: r,
		s:      s,
	}
	return h, h.setup()
}

func (h *Handler) setup() error {
	auth := h.Group("/oauth")
	{
		auth.POST("/google", h.googleLogin)
	}

	return nil
}

func (h *Handler) googleLogin(c *gin.Context) {
	var req service.GoogleAuthRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	res, err := h.s.GoogleAuth(c, req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
