package main

import (
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/config"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/handler"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {

	// обрабатываем аргументы командной строки
	config, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	// Инициализируем логер
	lgr, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log = lgr
	logger.Log.Infow("starting server",
		"address", config.AddrServer,
		"logLevel", config.LogLevel)
	store := repository.New()

	r := chi.NewRouter()
	r.Use(logger.LogRequest, logger.LogResponse)
	r.Post("/update/{type}/{name}/{value}", handler.UpdatePage(store))
	r.Get("/value/{type}/{name}", handler.GetCountMetric(store))
	r.Get("/", handler.AllMetrics(store))
	r.Post("/update/", handler.UpdateJSON(store))
	r.Post("/value/", handler.ValueJSON(store))

	if err := runServer(r, config); err != nil {
		log.Fatal(err)
	}
}

func runServer(router chi.Router, cfg *config.ServerCfg) error {
	/*if err := logger.Initialize(cfg.LogLevel); err != nil {
		return err
	}
	logger.Log.Info("Running server", zap.String("address", cfg.AddrServer))*/
	//fmt.Println("Starting server", cfg.AddrServer)
	return http.ListenAndServe(cfg.AddrServer, router)
}
