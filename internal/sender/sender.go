package sender

import (
	"fmt"
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
