package ws

import "time"

var (
	defaultMaxConnections     = 1000
	defaultMaxUserConnections = 3
	defaultPingInterval       = 30 * time.Second
)

type WsConfigs struct {
	MaxConnections     int
	MaxUserConnections int
	PingInterval       time.Duration
}

var defaultConfigs = &WsConfigs{
	MaxConnections:     defaultMaxConnections,
	MaxUserConnections: defaultMaxUserConnections,
	PingInterval:       defaultPingInterval,
}
