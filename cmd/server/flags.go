package main

import (
	"flag"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/config"
	"os"
	"strconv"
)

func parseFlags() (*config.ServerCfg, error) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	options := []config.ServerOption{
		addr(fs),
		logLvl(fs),
		storeInterval(fs),
		fileStorage(fs),
		restore(fs),
		databaseDSN(fs),
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
		/*if env := os.Getenv("ADDRESS"); env != "" {
			cfg.AddrServer = env
			return
		}*/
		if val, ok := os.LookupEnv("ADDRESS"); ok {
			cfg.AddrServer = val
			return
		}
		cfg.AddrServer = addrFlag
	}
}

func logLvl(fs *flag.FlagSet) config.ServerOption {
	var lvlFlag string
	fs.StringVar(&lvlFlag, "l", "info", "log level")

	return func(cfg *config.ServerCfg) {
		/*if env := os.Getenv("LOG_LEVEL"); env != "" {
			cfg.LogLevel = env
			return
		}*/
		if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
			cfg.LogLevel = val
			return
		}
		cfg.LogLevel = lvlFlag
	}
}

func storeInterval(fs *flag.FlagSet) config.ServerOption {
	var storeIntervalFlag int
	fs.IntVar(&storeIntervalFlag, "i", 2, "period of save metrics. 0 is sync mode")

	return func(cfg *config.ServerCfg) {
		/*if env := os.Getenv("STORE_INTERVAL"); env != "" {
			if v, err := strconv.Atoi(env); err == nil {
				cfg.StoreInterval = v
				return
			}
		}*/
		if val, ok := os.LookupEnv("STORE_INTERVAL"); ok {
			cfg.StoreInterval, _ = strconv.Atoi(val)
			return
		}
		cfg.StoreInterval = storeIntervalFlag
	}
}

func fileStorage(fs *flag.FlagSet) config.ServerOption {
	var fileStorageFlag string
	fs.StringVar(&fileStorageFlag, "f", "" /*"storage.txt"*/, "path to file storage to use")

	return func(cfg *config.ServerCfg) {
		/*if env := os.Getenv("FILE_STORAGE_PATH"); env != "" {
			cfg.FileStorage = env
			return
		}*/
		if val, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
			cfg.FileStorage = val
			return
		}
		cfg.FileStorage = fileStorageFlag
	}
}

func restore(fs *flag.FlagSet) config.ServerOption {
	var restoreFlag bool
	fs.BoolVar(&restoreFlag, "r", false, "need to restore metrics")

	return func(cfg *config.ServerCfg) {
		/*if env := os.Getenv("RESTORE"); env != "" {
			if v, err := strconv.ParseBool(env); err == nil {
				cfg.Restore = v
				return
			}
			return
		}*/
		if val, ok := os.LookupEnv("RESTORE"); ok {
			if v, err := strconv.ParseBool(val); err == nil {
				cfg.Restore = v
				return
			}
		}

		cfg.Restore = restoreFlag
	}
}

func databaseDSN(fs *flag.FlagSet) config.ServerOption {
	var databaseFlag string
	fs.StringVar(&databaseFlag, "d",
		//"host=localhost user=postgres password=1Qaz2wsx dbname=videos sslmode=disable",
		"",
		"connection string for database")

	return func(cfg *config.ServerCfg) {
		/*if env := os.Getenv("DATABASE_DSN"); env != "" {
			cfg.Database = env
			return
		}*/
		if val, ok := os.LookupEnv("DATABASE_DSN"); ok {
			cfg.Database = val
			return
		}
		cfg.Database = databaseFlag
	}
}
