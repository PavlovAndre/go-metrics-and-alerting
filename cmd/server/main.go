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
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	// обрабатываем аргументы  командной строки
	config, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	// Канал для сигналов
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

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

	if config.Restore {
		fileStore.Read(config.FileStorage)
	}
	r := chi.NewRouter()

	r.Use(logger.LogRequest, logger.LogResponse, compress.GzipMiddleware)
	r.Post("/update/{type}/{name}/{value}", handler.UpdatePage(store))
	r.Get("/value/{type}/{name}", handler.GetCountMetric(store))
	r.Get("/", handler.AllMetrics(store))
	r.Post("/update/", handler.UpdateJSON(store))
	r.Post("/value/", handler.ValueJSON(store))

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		fileStore.Write(config.FileStorage)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := runServer(r, config); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()
		// Ожидание сигнала
		<-quit
		log.Println("Получен сигнал завершения")
		fileStore.WriteEnd(config.FileStorage)
		os.Exit(0)
	}()

	wg.Wait()
	logger.Log.Infow("server stopped")

}

func runServer(router chi.Router, cfg *config.ServerCfg) error {
	return http.ListenAndServe(cfg.AddrServer, router)
}
