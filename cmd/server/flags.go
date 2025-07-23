package main

import (
	"flag"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/config"
	"os"
)

func parseFlags() (*config.ServerCfg, error) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	options := []config.ServerOption{
		addr(fs),
		logLvl(fs),
	}

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}
	return config.NewServerCfg(options...), nil
}

func addr(fs *flag.FlagSet) config.ServerOption {
	var addrFlag string
	fs.StringVar(&addrFlag, "a", "localhost:8080", "address and port to run server")

	return func(cfg *config.ServerCfg) {
		if env := os.Getenv("ADDRESS"); env != "" {
			cfg.AddrServer = env
			return
		}
		cfg.AddrServer = addrFlag
	}
}

func logLvl(fs *flag.FlagSet) config.ServerOption {
	var lvlFlag string
	fs.StringVar(&lvlFlag, "l", "info", "log level")

	return func(cfg *config.ServerCfg) {
		if env := os.Getenv("LOG_LEVEL"); env != "" {
			cfg.LogLevel = env
			return
		}
		cfg.LogLevel = lvlFlag
	}
}
