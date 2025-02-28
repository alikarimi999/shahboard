package authservice

import (
	"github.com/alikarimi999/shahboard/authservice/deliver/http"
	"github.com/alikarimi999/shahboard/authservice/repository"
	auth "github.com/alikarimi999/shahboard/authservice/service"
	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/pkg/postgres"
)

type application struct {
	*auth.AuthService
	*http.Handler
}

func SetupApplication(cfg Config) (*application, error) {
	p, _, err := kafka.NewKafkaPublisherAndSubscriber(cfg.Kafka, log.NewLogger(cfg.Log.File, cfg.Log.Verbose))
	if err != nil {
		return nil, err
	}

	db, err := postgres.Setup(cfg.PostgresDB)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := repository.NewUserRepo(db)

	jwtGenerator, err := jwt.NewGenerator(cfg.JwtGenerator)
	if err != nil {
		return nil, err
	}

	l := log.NewLogger(cfg.Log.File, cfg.Log.Verbose)

	svc := auth.NewAuthService(cfg.Auth, repo, jwtGenerator, p, l)

	handler, err := http.NewHandler(cfg.Http, svc)
	if err != nil {
		return nil, err
	}

	return &application{
		AuthService: svc,
		Handler:     handler,
	}, nil
}

func (a *application) Run() error {
	return a.Router.Run()
}
