package sender

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/compress"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
)

type Result struct {
	Value    any
	Err      error
	WorkerID int
}
type WorkerPool struct {
	JobsCh     chan *Work
	addrServer string
	hashKey    string
	closed     atomic.Bool
}

type Work struct {
	Metrics []models.Metrics
	Result  chan error
}

func NewWorkerPool(workers int, addr string, key string) *WorkerPool {
	w := &WorkerPool{
		JobsCh:     make(chan *Work),
		addrServer: addr,
		hashKey:    key,
	}
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := range workers {
		go w.RunWorker(&wg, i)
	}
	return w
}

/*func (wp *WorkerPool) Stop() {
	if wp.closed.Load() {
		return
	}
	wp.closed.Store(true)
	close(wp.JobsCh)
}*/

func (wp *WorkerPool) Send(metrics []models.Metrics) error {
	work := Work{}
	work.Metrics = metrics
	work.Result = make(chan error)
	defer close(work.Result)
	wp.JobsCh <- &work
	return <-work.Result
}

func (wp *WorkerPool) SendWorkerPool(metrics []models.Metrics) error {
	log.Printf("Start func SendMetrics2")
	client := &http.Client{}
	body, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error marshalling json: %s\n", err)
		return err
	}
	compressBody, err := compress.GZIPCompress(body)
	if err != nil {
		log.Printf("Error compressing json: %s\n", err)
	}

	sendURL := fmt.Sprintf("http://%s/updates/", wp.addrServer)
	conn, err := net.DialTimeout("tcp", wp.addrServer, 0)
	if err != nil {
		log.Printf("Error connecting to %s: %s\n", sendURL, err)
		return err
	}
	conn.Close()

	req, err := http.NewRequest("POST", sendURL, bytes.NewReader(compressBody))
	if err != nil {
		log.Printf("ошибка создания запроса")
		return err
	}
	// Устанавливаем заголовок
	if wp.hashKey != "" {
		bodyHash := wp.hashBody(body)
		req.Header.Set("HashSHA256", bodyHash)
		log.Printf("Установили заголовок HashSHA256 %s", bodyHash)
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ошибка отправки запроса %s", err)
		return err
	}
	resp.Body.Close()
	return nil

}

func (wp *WorkerPool) RunWorker(wg *sync.WaitGroup, id int) {
	log.Printf("Start func RunWorker %d", id)
	defer wg.Done()
	/*for {
		v, ok := <-wp.JobsCh
		if !ok {
			log.Printf("jobs channel closed, closing worker")
			return
		}
		log.Printf("worker received job")

		err := wp.SendWorkerPool(v.Metrics)
		v.Result <- err
	}*/
	for v := range wp.JobsCh {
		log.Printf("worker received job")

		err := wp.SendWorkerPool(v.Metrics)
		v.Result <- err
	}
}

// hashBody создаём подпись запроса
func (wp *WorkerPool) hashBody(body []byte) string {
	harsher := hmac.New(sha256.New, []byte(wp.hashKey))
	harsher.Write(body)
	return hex.EncodeToString(harsher.Sum(nil))
}
