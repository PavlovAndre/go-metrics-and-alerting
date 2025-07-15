package main

import (
	"flag"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/config"
	"os"
	"strconv"
)

// Парсим командную строку, получаем адрес сервера, интервалы сбора и отправки метрик
func parseFlags() (*config.AgentCfg, error) {
	fs := flag.NewFlagSet("agent", flag.ExitOnError)
	options := []config.AgentOption{
		flagAddr(fs),
		flagPollInterval(fs),
		flagReportInterval(fs),
	}
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}
	return config.NewAgentCfg(options...), nil
}

func flagAddr(fs *flag.FlagSet) config.AgentOption {
	var addrFlag string
	fs.StringVar(&addrFlag, "a", "localhost:8080", "address server")

	return func(cfg *config.AgentCfg) {
		if env := os.Getenv("ADDRESS"); env != "" {
			cfg.AddrServer = env
			return
		}
		cfg.AddrServer = addrFlag
	}
}

func flagPollInterval(fs *flag.FlagSet) config.AgentOption {
	var pollInterval int
	fs.IntVar(&pollInterval, "p", 2, "poll interval")

	return func(cfg *config.AgentCfg) {
		if env := os.Getenv("POLL_INTERVAL"); env != "" {
			if v, err := strconv.Atoi(env); err == nil {
				cfg.PollInterval = v
				return
			}
		}
		cfg.PollInterval = pollInterval
	}
}

func flagReportInterval(fs *flag.FlagSet) config.AgentOption {
	var reportInterval int
	fs.IntVar(&reportInterval, "r", 10, "report interval")

	return func(cfg *config.AgentCfg) {
		if env := os.Getenv("REPORT_INTERVAL"); env != "" {
			if v, err := strconv.Atoi(env); err == nil {
				cfg.ReportInterval = v
				return
			}
		}
		cfg.ReportInterval = reportInterval
	}
}
