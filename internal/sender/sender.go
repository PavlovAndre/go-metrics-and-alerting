package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/compress"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/metricserror"
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
			log.Printf("Start func SendMetricsJSON")
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
				//req.Header.Set("Content-Type", "application/json")
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
				//req.Header.Set("Content-Type", "application/json")
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

// SendMetricsBatchJSON Функция отправки  метрик по JSON одним батчем
func (s *Sender) SendMetricsBatchJSON() {
	for {
		ticker := time.NewTicker(time.Duration(s.reportInterval) * time.Second)
		for range ticker.C {
			log.Printf("Start func SendMetricsBatchJSON")
			var metrics []models.Metrics
			for key, value := range s.memStore.GetGauges() {
				send := models.Metrics{
					ID:    key,
					MType: "gauge",
					Value: &value,
				}
				metrics = append(metrics, send)
			}
			//log.Printf("Gauges %v", metrics)
			for key, value := range s.memStore.GetCounters() {
				send := models.Metrics{
					ID:    key,
					MType: "counter",
					Delta: &value,
				}
				metrics = append(metrics, send)
			}
			//log.Printf("Counters %v", metrics)
			body, err := json.Marshal(metrics)
			if err != nil {
				log.Printf("Error marshalling json: %s\n", err)
				continue
			}
			//log.Printf("Body %v", body)
			compressBody, err := compress.GZIPCompress(body)
			if err != nil {
				log.Printf("Error compressing json: %s\n", err)
			}

			sendURL := fmt.Sprintf("http://%s/updates/", s.addrServer)
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
			//req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding", "gzip")
			req.Header.Set("Accept-Encoding", "gzip")
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("ошибка отправки запроса")
				continue
			}
			resp.Body.Close()
			s.memStore.SetCounter("PollCount", 0)
		}
	}
}

// SendMetricsBatchJSONPeriod Функция отправки  метрик по JSON одним батчем
func (s *Sender) SendMetricsBatchJSONPeriod(ctx context.Context) {
	logger.Log.Info("Starting periodic sender")
	ticker := time.NewTicker(time.Duration(s.reportInterval) * time.Second)
	s.retrySend()
	for {
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-ticker.C:
			s.retrySend()
		case <-ctx.Done():
			logger.Log.Info("Periodic sender stopped")
			ticker.Stop()
			return
		}
	}

}

// retrySend отправка метрик с повторами
func (s *Sender) retrySend() {
	pause := time.Second
	var rErr *metricserror.Retriable
	for i := 0; i < 3; i++ {
		log.Printf("Starting retry send %v", i)
		err := s.SendMetrics2()
		log.Printf("Starting retry send2 %v", i)
		if err == nil {
			log.Printf("Starting retry send3 %v", i)
			break
		}
		log.Printf("Starting retry send4 %v", i)
		logger.Log.Error(err)
		log.Printf("Starting retry send5 %v", i)
		if !errors.As(err, &rErr) {
			break
		}
		log.Printf("Starting retry send6 %v", i)
		<-time.After(pause)
		pause += 2 * time.Second
	}
}

func (s *Sender) SendMetrics2() error {
	log.Printf("Start func SendMetricsBatchJSONPeriod")
	s.memStore.Mu.RLock()
	defer s.memStore.Mu.RUnlock()
	var metrics []models.Metrics
	for key, value := range s.memStore.GetGauges() {
		send := models.Metrics{
			ID:    key,
			MType: "gauge",
			Value: &value,
		}
		metrics = append(metrics, send)
	}
	//log.Printf("Gauges %v", metrics)
	for key, value := range s.memStore.GetCounters() {
		send := models.Metrics{
			ID:    key,
			MType: "counter",
			Delta: &value,
		}
		metrics = append(metrics, send)
	}
	//log.Printf("Counters %v", metrics)
	body, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error marshalling json: %s\n", err)
		return err
	}
	//log.Printf("Body %v", body)
	compressBody, err := compress.GZIPCompress(body)
	if err != nil {
		log.Printf("Error compressing json: %s\n", err)
	}

	sendURL := fmt.Sprintf("http://%s/updates/", s.addrServer)
	conn, err := net.DialTimeout("tcp", s.addrServer, 0)
	if err != nil {
		log.Printf("Error connecting to %s: %s\n", sendURL, err)
		return err
	}
	conn.Close()

	client := &http.Client{}

	req, err := http.NewRequest("POST", sendURL, bytes.NewReader(compressBody))
	if err != nil {
		log.Printf("ошибка создания запроса")
		return err
	}
	//req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ошибка отправки запроса")
		return err
	}
	resp.Body.Close()
	s.memStore.SetCounter("PollCount", 0)
	return nil
}
