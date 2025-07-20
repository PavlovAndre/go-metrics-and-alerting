package config

type ServerCfg struct {
	AddrServer string
	LogLevel   string
}

type ServerOption func(*ServerCfg)

func NewServerCfg(opts ...ServerOption) *ServerCfg {
	cfg := &ServerCfg{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
