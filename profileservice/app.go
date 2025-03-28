package profileservice

import (
	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/pkg/postgres"
	"github.com/alikarimi999/shahboard/profileservice/delivery/grpc"
	"github.com/alikarimi999/shahboard/profileservice/delivery/http"
	"github.com/alikarimi999/shahboard/profileservice/repository"
	"github.com/alikarimi999/shahboard/profileservice/service/rating"
	"github.com/alikarimi999/shahboard/profileservice/service/user"
)

type application struct {
	rating *rating.Service
	user   *user.Service
	http   *http.Handler
	grpc   *grpc.Server
	l      log.Logger
}

func SetupApplication(cfg Config) (*application, error) {

	l := log.NewLogger(cfg.Log.File, cfg.Log.Verbose)

	_, s, err := kafka.NewKafkaPublisherAndSubscriber(cfg.Kafka, l)
	if err != nil {
		return nil, err
	}

	userDB, err := postgres.Setup(cfg.UsersDB)
	if err != nil {
		return nil, err
	}

	ratingDB, err := postgres.Setup(cfg.RatingDB)
	if err != nil {
		return nil, err
	}

	userRepo := repository.NewUserRepo(userDB)
	ratingRepo := repository.NewRatingRepo(ratingDB)

	userService := user.NewService(cfg.User, userRepo, s, l)
	ratingService := rating.NewService(cfg.Rating, ratingRepo, s, l)

	h, err := http.NewHandler(cfg.Http, userService, ratingService, l)
	if err != nil {
		return nil, err
	}

	grpcServer, err := grpc.NewServer(cfg.Grpc, ratingService)
	if err != nil {
		return nil, err
	}

	return &application{
		rating: ratingService,
		user:   userService,
		http:   h,
		grpc:   grpcServer,
	}, nil
}

func (a *application) Run() error {
	go func() {
		if err := a.grpc.Run(); err != nil {
			a.l.Fatal(err.Error())
		}
	}()

	if err := a.http.Run(); err != nil {
		return err
	}

	return nil
}
