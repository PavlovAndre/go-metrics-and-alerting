package config

type AgentCfg struct {
	AddrServer     string
	PollInterval   int
	ReportInterval int
}

type AgentOption func(*AgentCfg)

func NewAgentCfg(opts ...AgentOption) *AgentCfg {
	cfg := &AgentCfg{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
