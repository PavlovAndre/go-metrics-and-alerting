package main

import (
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/collector"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/sender"
	"log"
	"sync"
)

func main() {
	logger.Log.Infow("Starting agent")
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
		logger.Log.Infow("Starting collector")
		defer wg.Done()
		coll.CollectMetrics()
	}()
	go func() {
		logger.Log.Infow("Starting sender")
		defer wg.Done()
		send.SendMetricsJSON()
	}()
	wg.Wait()

}
