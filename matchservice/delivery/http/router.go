package http

import (
	"fmt"

	match "github.com/alikarimi999/shahboard/matchservice/service"
	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Port int `json:"port"`
}

type Router struct {
	cfg Config
	gin *gin.Engine
	s   *match.Service
	v   *jwt.Validator
}

func NewRouter(cfg Config, s *match.Service, v *jwt.Validator) (*Router, error) {
	// gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(middleware.Cors())

	r := &Router{
		cfg: cfg,
		gin: engine,
		s:   s,
		v:   v,
	}

	return r, r.setup()
}

func (r *Router) Run() error {
	return r.gin.Run(fmt.Sprintf(":%d", r.cfg.Port))
}

func (r *Router) setup() error {
	r.setupUserRoutes()

	return nil
}
