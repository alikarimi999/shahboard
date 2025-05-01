package http

import (
	"net/http"

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
	h.Handle(http.MethodPost, "/", h.passwordLogin)
	h.Handle(http.MethodGet, "/guest", h.guestLogin)
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

func (h *Handler) passwordLogin(c *gin.Context) {
	var req service.PasswordAuthRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	res, err := h.s.PasswordAuth(c, req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func (h *Handler) guestLogin(c *gin.Context) {
	res, err := h.s.GuestLogin(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
