package main

import (
	"flag"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/config"
	"os"
)

/*
var flagRunAddr string

	func parseFlags() {
		flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
		flag.Parse()
	}
*/
func parseFlags() (*config.ServerCfg, error) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	options := []config.ServerOption{
		addr(fs),
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
