package main

import (
	"context"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/collector"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/sender"
	"log"
	"sync"
)

func main() {
	logger.Log.Infow("Starting agent")
	configAgent, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(2)
	store := repository.New()
	coll := collector.New(store, configAgent.PollInterval)
	send := sender.New(store, configAgent.ReportInterval, configAgent.AddrServer, configAgent.HashKey)

	go func() {
		logger.Log.Infow("Starting collector")
		defer wg.Done()
		coll.CollectMetrics()
	}()
	go func() {
		logger.Log.Infow("Starting system collector")
		defer wg.Done()
		coll.CollectSystemMetrics()
	}()
	go func() {
		logger.Log.Infow("Starting sender")
		defer wg.Done()
		send.SendMetricsBatchJSONPeriod(ctx)
	}()
	wg.Wait()

}
