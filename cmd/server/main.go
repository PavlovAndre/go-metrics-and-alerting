package main

import (
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

	/*//Контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()*/

	logger.Log = lgr
	logger.Log.Infow("starting server",
		"address", config.AddrServer,
		"logLevel", config.LogLevel,
		"path", config.FileStorage,
		"restore", config.Restore,
		"storeInterval", config.StoreInterval,
		"database", config.Database)
	store := repository.New()

	fileStore := logger.NewFileStorage(store, config.StoreInterval)

	//Подключение к базе
	ps := config.Database
	db, err := sql.Open("pgx", ps)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	logger.Log.Info(ps)
	if config.Database != "" {
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

	if config.Restore {
		fileStore.Read(config.FileStorage)
	}
	r := chi.NewRouter()

	r.Use(logger.LogRequest, logger.LogResponse, compress.GzipMiddleware)
	r.Group(func(r1 chi.Router) {
		r1.Post("/update/{type}/{name}/{value}", handler.UpdatePage(store))
		r1.Get("/value/{type}/{name}", handler.GetCountMetric(store))
	})
	if config.Database != "" {
		r.Get("/", handler.AllDB(db))
		r.Group(func(r2 chi.Router) {
			//r2.Use(handler.SetContentType)
			r2.Post("/update/", handler.UpdateDB(db))
			r2.Post("/value/", handler.ValueDB(db))
			r2.Post("/updates/", handler.UpdatesDB(db))
		})

	} else {
		r.Get("/", handler.AllMetrics(store))
		r.Group(func(r2 chi.Router) {
			//r2.Use(handler.SetContentType)
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
		if config.FileStorage != "" {
			fileStore.Write(config.FileStorage)
			if err != nil {
				log.Fatal(err)
			}
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
		if config.FileStorage != "" {
			fileStore.WriteEnd(config.FileStorage)
		}
		os.Exit(0)
	}()

	wg.Wait()
	logger.Log.Infow("server stopped")

}

func runServer(router chi.Router, cfg *config.ServerCfg) error {
	return http.ListenAndServe(cfg.AddrServer, router)
}
