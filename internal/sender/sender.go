package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"log"
	"net/http"
	"time"
)

type Sender struct {
	memStore       *repository.MemStore
	reportInterval int
	addrServer     string
}

func New(store *repository.MemStore, sendInt int, addr string) *Sender {
	return &Sender{memStore: store, reportInterval: sendInt, addrServer: addr}
}

// SendMetrics Функция отправки метрик
func (s *Sender) SendMetrics() {
	for {
		ticker := time.NewTicker(time.Duration(s.reportInterval) * time.Second)
		for range ticker.C {
			for key, value := range s.memStore.GetGauges() {
				sendURL := fmt.Sprintf("http://%s/update/%s/%s/%f", s.addrServer, "gauge", key, value)
				resp, err := http.Post(sendURL, "text/plain", nil)
				resp.Body.Close()
				if err != nil {
					log.Printf("Error posting to %s: %s\n", sendURL, err)
				}
				fmt.Println(resp)
			}
			for key, value := range s.memStore.GetCounters() {
				sendURL := fmt.Sprintf("http://%s/update/%s/%s/%d", s.addrServer, "counter", key, value)
				resp, err := http.Post(sendURL, "text/plain", nil)
				resp.Body.Close()
				if err != nil {
					log.Printf("Error posting to %s: %s\n", sendURL, err)
				}
				fmt.Println(resp)

			}
			s.memStore.SetCounter("PollCount", 0)
		}
	}
}

// SendMetricsJSON Функция отправки метрик по  JSON
func (s *Sender) SendMetricsJSON() {
	for {
		ticker := time.NewTicker(time.Duration(s.reportInterval) * time.Second)
		for range ticker.C {
			for key, value := range s.memStore.GetGauges() {
				send := models.Metrics{
					ID:    key,
					MType: "gauge",
					Value: &value,
				}
				body, err := json.Marshal(send)
				if err != nil {
					log.Printf("Error marshalling json: %s\n", err)
					continue
				}
				sendURL := fmt.Sprintf("http://%s/update/", s.addrServer)
				resp, err := http.Post(sendURL, "application/json", bytes.NewReader(body))
				resp.Body.Close()
				if err != nil {
					log.Printf("Error posting to %s: %s\n", sendURL, err)
				}
				fmt.Println(resp)
			}
			for key, value := range s.memStore.GetCounters() {
				send := models.Metrics{
					ID:    key,
					MType: "counter",
					Delta: &value,
				}
				body, err := json.Marshal(send)
				if err != nil {
					log.Printf("Error marshalling json: %s\n", err)
					continue
				}
				sendURL := fmt.Sprintf("http://%s/update/", s.addrServer)
				resp, err := http.Post(sendURL, "application/json", bytes.NewReader(body))
				resp.Body.Close()
				if err != nil {
					log.Printf("Error posting to %s: %s\n", sendURL, err)
				}
				fmt.Println(resp)

			}
			s.memStore.SetCounter("PollCount", 0)
		}
	}
}
