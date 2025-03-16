package ws

var (
	defaultMaxConnections     = 1000
	defaultMaxUserConnections = 3
	defaultUserSessionsCap    = 3
)

type WsConfigs struct {
	MaxConnections     int
	MaxUserConnections int
	UserSessionsCap    int
}

var defaultConfigs = &WsConfigs{
	MaxConnections:     defaultMaxConnections,
	MaxUserConnections: defaultMaxUserConnections,
	UserSessionsCap:    defaultUserSessionsCap,
}
