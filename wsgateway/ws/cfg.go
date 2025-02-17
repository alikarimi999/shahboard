package ws

import "time"

var (
	defaultMaxConnections                  = 1000
	defaultMaxUserConnections              = 3
	defaultPingIntervalDeadSession         = 1 * time.Minute
	defaultPingIntervalDisconnectedSession = 5 * time.Second
)

type WsConfigs struct {
	MaxConnections                  int
	MaxUserConnections              int
	PingIntervalDeadSession         time.Duration
	PingIntervalDisconnectedSession time.Duration
}

var defaultConfigs = &WsConfigs{
	MaxConnections:                  defaultMaxConnections,
	MaxUserConnections:              defaultMaxUserConnections,
	PingIntervalDeadSession:         defaultPingIntervalDeadSession,
	PingIntervalDisconnectedSession: defaultPingIntervalDisconnectedSession,
}
