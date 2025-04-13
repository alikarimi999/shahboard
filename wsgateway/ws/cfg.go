package ws

var (
	defaultMaxConnections  = 1000
	defaultUserSessionsCap = 3
)

type WsConfigs struct {
	MaxConnections  int `json:"max_connections"`
	UserSessionsCap int `json:"user_sessions_cap"`
}

var defaultConfigs = &WsConfigs{
	MaxConnections:  defaultMaxConnections,
	UserSessionsCap: defaultUserSessionsCap,
}
