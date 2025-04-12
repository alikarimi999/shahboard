package ws

var (
	defaultMaxConnections  = 1000
	defaultUserSessionsCap = 3
)

type WsConfigs struct {
	MaxConnections  int
	UserSessionsCap int
}

var defaultConfigs = &WsConfigs{
	MaxConnections:  defaultMaxConnections,
	UserSessionsCap: defaultUserSessionsCap,
}
