package main

import (
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/handler"
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

	//repository.Store = repository.New()
	store := repository.New()

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.UpdatePage(store))
	r.Get("/value/{type}/{name}", handler.GetCountMetric(store))
	r.Get("/", handler.AllMetrics(store))

	if err := runServer(r, config.AddrServer); err != nil {
		log.Fatal(err)
	}
}

func runServer(router chi.Router, addr string) error {
	fmt.Println("Starting server", addr)
	return http.ListenAndServe(addr, router)
}
