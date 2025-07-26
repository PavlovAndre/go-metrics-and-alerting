package main

import (
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/compress"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/config"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/handler"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"sync"
)

func main() {

	// обрабатываем аргументы  командной строки
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
		"logLevel", config.LogLevel,
		"path", config.FileStorage,
		"restore", config.Restore,
		"storeInterval", config.StoreInterval)
	store := repository.New()
	fileStore := logger.NewFileStorage(store, config.StoreInterval)

	/*settings := logger.FileStorage{
		Port: 4000,
		Host: `localhost`,
	}*/
	//if err := settings.Save("test.txt" /*config.FileStorage*/); err != nil {
	//	logger.Log.Fatal(err)
	//}

	r := chi.NewRouter()
	//r2 := chi.NewRouter()
	//r.Use(logger.LogRequest, logger.LogResponse /*, compress.GzipMiddleware*/)
	r.Use(logger.LogRequest, logger.LogResponse, compress.GzipMiddleware)
	r.Post("/update/{type}/{name}/{value}", handler.UpdatePage(store))
	r.Get("/value/{type}/{name}", handler.GetCountMetric(store))
	r.Get("/", handler.AllMetrics(store))
	r.Post("/update/", handler.UpdateJSON(store))
	r.Post("/value/", handler.ValueJSON(store))

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		fileStore.Write(config.FileStorage)
		if err != nil {
			log.Fatal(err)
		}
	}()

	//go func() {
	//	defer wg.Done()
	if err := runServer(r, config); err != nil {
		log.Fatal(err)
	}
	//}()
	wg.Wait()
}

func runServer(router chi.Router, cfg *config.ServerCfg) error {
	/*if err := logger.Initialize(cfg.LogLevel); err != nil {
		return err
	}
	logger.Log.Info("Running server", zap.String("address", cfg.AddrServer))*/
	//fmt.Println("Starting server", cfg.AddrServer)
	return http.ListenAndServe(cfg.AddrServer, router)
}
