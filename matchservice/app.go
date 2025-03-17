package matchservice

import (
	"math/rand"

	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/matchservice/delivery/http"
	match "github.com/alikarimi999/shahboard/matchservice/service"
	"github.com/alikarimi999/shahboard/matchservice/services/game"

	"github.com/alikarimi999/shahboard/pkg/grpc"
	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
)

type application struct {
	*match.Service
	*http.Router
}

func SetupApplication(cfg Config) (*application, error) {

	l := log.NewLogger(cfg.Log.File, cfg.Log.Verbose)

	v, err := jwt.NewValidator(cfg.JwtValidator)
	if err != nil {
		return nil, err
	}

	p, _, err := kafka.NewKafkaPublisherAndSubscriber(cfg.Kafka, l)
	if err != nil {
		return nil, err
	}

	client, err := grpc.NewClient(cfg.GameService, nil)
	if err != nil {
		return nil, err
	}

	s, err := match.NewService(cfg.Match, p, &mockScoreService{}, game.NewService(client), l)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRouter(cfg.Http, s, v)
	if err != nil {
		return nil, err
	}

	return &application{Service: s, Router: r}, nil
}

func (a *application) Run() error {
	return a.Router.Run()
}

type mockScoreService struct{}

func (s *mockScoreService) GetUserLevel(id types.ObjectId) (types.Level, error) {
	return types.Level(rand.Intn(int(types.LevelKing)) + 1), nil
}
