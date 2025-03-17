package grpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var DefaultDialOption = grpc.WithTransportCredentials(insecure.NewCredentials())

type Config struct {
	Target string `json:"target"`
}

// NewClient returns a new grpc client connection.
// If option is nil, DefaultDialOption will be used.
func NewClient(cfg Config, option grpc.DialOption) (*grpc.ClientConn, error) {
	if cfg.Target == "" {
		return nil, fmt.Errorf("grpc target is required")
	}

	if option == nil {
		option = DefaultDialOption
	}

	conn, err := grpc.NewClient(cfg.Target, option)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
