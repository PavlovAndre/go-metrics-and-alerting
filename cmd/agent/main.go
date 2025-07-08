package main

import (
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Функция сбора метрик
func collectMetrics(pollInterval int) {
	for {
		m := runtime.MemStats{}
		runtime.ReadMemStats(&m)
		repository.Store.SetGauge("Alloc", float64(m.Alloc))
		repository.Store.SetGauge("BuckHashSys", float64(m.BuckHashSys))
		repository.Store.SetGauge("Frees", float64(m.Frees))
		repository.Store.SetGauge("GCCPUFraction", float64(m.GCCPUFraction))
		repository.Store.SetGauge("GCSys", float64(m.GCSys))
		repository.Store.SetGauge("HeapAlloc", float64(m.HeapAlloc))
		repository.Store.SetGauge("HeapIdle", float64(m.HeapIdle))
		repository.Store.SetGauge("HeapInuse", float64(m.HeapInuse))
		repository.Store.SetGauge("HeapObjects", float64(m.HeapObjects))
		repository.Store.SetGauge("HeapReleased", float64(m.HeapReleased))
		repository.Store.SetGauge("HeapSys", float64(m.HeapSys))
		repository.Store.SetGauge("LastGC", float64(m.LastGC))
		repository.Store.SetGauge("Lookups", float64(m.Lookups))
		repository.Store.SetGauge("MCacheInuse", float64(m.MCacheInuse))
		repository.Store.SetGauge("MCacheSys", float64(m.MCacheSys))
		repository.Store.SetGauge("MSpanInuse", float64(m.MSpanInuse))
		repository.Store.SetGauge("MSpanSys", float64(m.MSpanSys))
		repository.Store.SetGauge("Mallocs", float64(m.Mallocs))
		repository.Store.SetGauge("NextGC", float64(m.NextGC))
		repository.Store.SetGauge("NumForcedGC", float64(m.NumForcedGC))
		repository.Store.SetGauge("NumGC", float64(m.NumGC))
		repository.Store.SetGauge("OtherSys", float64(m.OtherSys))
		repository.Store.SetGauge("PauseTotalNs", float64(m.PauseTotalNs))
		repository.Store.SetGauge("StackInuse", float64(m.StackInuse))
		repository.Store.SetGauge("StackSys", float64(m.StackSys))
		repository.Store.SetGauge("Sys", float64(m.Sys))
		repository.Store.SetGauge("TotalAlloc", float64(m.TotalAlloc))
		repository.Store.SetGauge("RandomValue", rand.Float64())
		repository.Store.SetCounter("PollCount", 1)

		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}

// Функция отправки метрик
func sendMetrics(addrServer string, reportInterval int) {
	for {
		for key, value := range repository.Store.GetGauges() {
			//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
			sendURL := fmt.Sprintf("http://%s/update/%s/%s/%f", addrServer, "gauge", key, value)
			resp, err := http.Post(sendURL, "text/plain", nil)
			resp.Body.Close()
			if err != nil {
				fmt.Printf("Error posting to %s: %s\n", sendURL, err)
			}
			fmt.Println(resp)

		}
		for key, value := range repository.Store.GetCounters() {
			//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>"localhost:8080"
			sendURL := fmt.Sprintf("http://%s/update/%s/%s/%d", addrServer, "counter", key, value)
			resp, err := http.Post(sendURL, "text/plain", nil)
			resp.Body.Close()
			if err != nil {
				fmt.Printf("Error posting to %s: %s\n", sendURL, err)
			}
			fmt.Println(resp)

		}
		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}

func main() {
	config, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	repository.Store = repository.New()
	go func() {
		defer wg.Done()
		collectMetrics(config.PollInterval)
	}()
	go func() {
		defer wg.Done()
		sendMetrics(config.AddrServer, config.ReportInterval)
	}()
	wg.Wait()

}
