package sender

import (
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
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
		for key, value := range s.memStore.GetGauges() {
			sendURL := fmt.Sprintf("http://%s/update/%s/%s/%f", s.addrServer, "gauge", key, value)
			resp, err := http.Post(sendURL, "text/plain", nil)
			resp.Body.Close()
			if err != nil {
				fmt.Printf("Error posting to %s: %s\n", sendURL, err)
			}
			fmt.Println(resp)
		}
		for key, value := range s.memStore.GetCounters() {
			sendURL := fmt.Sprintf("http://%s/update/%s/%s/%d", s.addrServer, "counter", key, value)
			resp, err := http.Post(sendURL, "text/plain", nil)
			resp.Body.Close()
			if err != nil {
				fmt.Printf("Error posting to %s: %s\n", sendURL, err)
			}
			fmt.Println(resp)

		}
		s.memStore.SetCounter("PollCount", 0)
		time.Sleep(time.Duration(s.reportInterval) * time.Second)
	}
}
