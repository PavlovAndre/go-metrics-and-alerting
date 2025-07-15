package main

import (
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/collector"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/sender"
	"log"
	"sync"
)

func main() {
	config, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	store := repository.New()
	coll := collector.New(store, config.PollInterval)
	send := sender.New(store, config.ReportInterval, config.AddrServer)

	go func() {
		defer wg.Done()
		coll.CollectMetrics()
	}()
	go func() {
		defer wg.Done()
		send.SendMetrics()
	}()
	wg.Wait()

}
