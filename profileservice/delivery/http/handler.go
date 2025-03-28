package http

import (
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/pkg/router"
	"github.com/alikarimi999/shahboard/profileservice/service/rating"
	"github.com/alikarimi999/shahboard/profileservice/service/user"
)

type Handler struct {
	*router.Router
	user   *user.Service
	rating *rating.Service
	l      log.Logger
}

func NewHandler(cfg router.Config, u *user.Service, r *rating.Service, l log.Logger) (*Handler, error) {
	router, err := router.NewRouter(cfg)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		Router: router,
		user:   u,
		rating: r,
		l:      l,
	}

	return h, h.setup()
}

func (h *Handler) Run() error {
	return h.Router.Run()
}

func (h *Handler) setup() error {
	h.setupUserRoutes()
	h.setupRatingRoutes()
	return nil
}
