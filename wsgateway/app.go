package wsgateway

import (
	"fmt"

	"github.com/alikarimi999/shahboard/event/kafka"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
	"github.com/gin-gonic/gin"
)

type application struct {
	server *ws.Server
	e      *gin.Engine
	cfg    Config
}

func SetupApplication(cfg Config) (*application, error) {

	l := log.NewLogger(cfg.Log.File, cfg.Log.Verbose)

	p, s, err := kafka.NewKafkaPublisherAndSubscriber(cfg.Kafka, l)
	if err != nil {
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)
	e := gin.Default()

	server, err := ws.NewServer(e, s, p, nil, l)
	if err != nil {
		return nil, err
	}

	return &application{
		server: server,
		e:      e,
		cfg:    cfg,
	}, nil
}

func (a *application) Start() error {
	return a.e.Run(fmt.Sprintf(":%s", a.cfg.Http.Port))
}
