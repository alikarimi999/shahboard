package router

import (
	"fmt"

	"github.com/alikarimi999/shahboard/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Port int `json:"port"`
}

type Router struct {
	cfg Config
	gin *gin.Engine
}

func NewRouter(cfg Config) (*Router, error) {
	// gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(middleware.Cors())

	r := &Router{
		cfg: cfg,
		gin: engine,
	}

	return r, nil
}

func (r *Router) Group(path string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return r.gin.Group(path, handlers...)
}

func (r *Router) Handle(method, path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.gin.Handle(method, path, handlers...)
}

func (r *Router) Run() error {
	return r.gin.Run(fmt.Sprintf(":%d", r.cfg.Port))
}
