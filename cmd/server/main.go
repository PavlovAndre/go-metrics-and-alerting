package main

import (
	"context"
	"database/sql"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/compress"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/config"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/handler"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/migrations"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func main() {

	// обрабатываем аргументы  командной строки
	configServer, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}
	config.Params = configServer

	// Контекст закрытия приложения
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Инициализируем логер
	lgr, err := logger.New(configServer.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log = lgr
	logger.Log.Infow("starting server",
		"address", configServer.AddrServer,
		"logLevel", configServer.LogLevel,
		"path", configServer.FileStorage,
		"restore", configServer.Restore,
		"storeInterval", configServer.StoreInterval,
		"database", configServer.Database,
		"hashKey", configServer.HashKey)
	store := repository.New()
	var fileStore *logger.FileStorage

	if configServer.Database == "" {
		fileStore = logger.NewFileStorage(store, configServer.StoreInterval)
		if configServer.Restore {
			fileStore.Read(configServer.FileStorage)
		}
	}

	//Подключение к базе
	ps := configServer.Database
	db, err := sql.Open("pgx", ps)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	logger.Log.Info(ps)
	if configServer.Database != "" {
		logger.Log.Info("Migrate migrations")
		// Применим миграции
		migrator, err := migrations.Migrations()
		if err != nil {
			logger.Log.Infow("failed to create migrations", "error", err)
		}
		if err = migrator.Migrate(db); err != nil {
			logger.Log.Infow("failed to migrate", "error", err)
		}
	}

	r := chi.NewRouter()

	r.Use(logger.LogRequest, logger.LogResponse, compress.GzipMiddleware, handler.CheckSign)
	r.Group(func(r1 chi.Router) {
		r1.Post("/update/{type}/{name}/{value}", handler.UpdatePage(store))
		r1.Get("/value/{type}/{name}", handler.GetCountMetric(store))
	})
	if configServer.Database != "" {
		r.Get("/", handler.AllDB(db))
		r.Group(func(r2 chi.Router) {
			//if configServer.HashKey != ""{r2.Use(handler.CheckSign)}
			r2.Post("/update/", handler.UpdateDB(db))
			r2.Post("/value/", handler.ValueDB(db))
			r2.Post("/updates/", handler.UpdatesDB(db))
		})

	} else {
		r.Get("/", handler.AllMetrics(store))
		r.Group(func(r2 chi.Router) {
			r2.Post("/update/", handler.UpdateJSON(store))
			r2.Post("/value/", handler.ValueJSON(store))
			r2.Post("/updates/", handler.UpdatesJSON(store))
		})
	}
	r.Get("/ping", handler.GetPing(db))

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		if configServer.FileStorage != "" {
			fileStore.Write(configServer.FileStorage)
			if err != nil {
				log.Fatal(err)
			}
		}

	}()

	go func() {
		defer wg.Done()
		if err := runServer(r, configServer); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()
		// Ожидание сигнала завершения
		<-ctx.Done()
		log.Println("Получен сигнал завершения")
		if configServer.FileStorage != "" {
			fileStore.WriteEnd(configServer.FileStorage)
		}
		os.Exit(0)
	}()

	wg.Wait()
	logger.Log.Infow("server stopped")

}

func runServer(router chi.Router, cfg *config.ServerCfg) error {
	return http.ListenAndServe(cfg.AddrServer, router)
}
