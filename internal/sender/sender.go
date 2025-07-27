package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/compress"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"log"
	"net"
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

// SendMetricsJSON Функция отправки  метрик по JSON
func (s *Sender) SendMetricsJSON() {
	for {
		ticker := time.NewTicker(time.Duration(s.reportInterval) * time.Second)
		for range ticker.C {
			log.Printf("Start func agent")
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

				compressBody, err := compress.GZIPCompress(body)
				if err != nil {
					log.Printf("Error compressing json: %s\n", err)
				}

				sendURL := fmt.Sprintf("http://%s/update/", s.addrServer)
				conn, err := net.DialTimeout("tcp", s.addrServer, 0)
				if err != nil {
					log.Printf("Error connecting to %s: %s\n", sendURL, err)
					continue
				}
				conn.Close()

				client := &http.Client{}

				req, err := http.NewRequest("POST", sendURL, bytes.NewReader(compressBody))
				if err != nil {
					log.Printf("ошибка создания запроса")
					continue
				}

				req.Header.Set("Content-Encoding", "gzip")
				req.Header.Set("Accept-Encoding", "gzip")
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("ошибка отправки запроса")
					continue
				}
				resp.Body.Close()
				//fmt.Println(resp)
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
				compressBody, err := compress.GZIPCompress(body)
				if err != nil {
					log.Printf("Error compressing json: %s\n", err)
				}

				sendURL := fmt.Sprintf("http://%s/update/", s.addrServer)
				conn, err := net.DialTimeout("tcp", s.addrServer, 0*time.Second)
				if err != nil {
					log.Printf("Error connecting to %s: %s\n", sendURL, err)
					continue
				}
				conn.Close()

				client := &http.Client{}
				req, err := http.NewRequest("POST", sendURL, bytes.NewReader(compressBody))
				if err != nil {
					log.Printf("ошибка создания запроса")
					continue
				}
				req.Body.Close()
				req.Header.Set("Content-Encoding", "gzip")
				req.Header.Set("Accept-Encoding", "gzip")
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("ошибка отправки запроса")
					continue
				}
				resp.Body.Close()
				fmt.Println(resp)
			}
			s.memStore.SetCounter("PollCount", 0)
		}
	}
}
