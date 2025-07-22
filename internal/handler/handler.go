package handler

import (
	"encoding/json"
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	memStore *repository.MemStore
	router   *chi.Mux
}

type metrics struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func New(store *repository.MemStore, root *chi.Mux) *Handler {
	return &Handler{memStore: store, router: root}
}

const templateHTML = `<!DOCTYPE html>
<html>
<body>
<h2>gauges</h2>
<ul>
{{range $key, $value := .Gauge}}<li>{{$key}} {{$value}}</li>{{end}}
</ul>
<h2>counters</h2>
<ul>
{{range $key, $value := .Counter}}<li>{{$key}} {{$value}}</li>{{end}}
</ul>
</body>
</html>
`

func UpdatePage(store *repository.MemStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Проверяем, что метод POST
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		// Проверям, что введен правильный тип метрик
		if metricType != "gauge" && metricType != "counter" {
			http.Error(w, "Bad type of metric", http.StatusBadRequest)
			return
		}

		//Проверка, что имя метрики не пустое
		if metricName == "" {
			http.NotFound(w, r)
			return
		}

		//Выполняем обновление значения gauge
		if metricType == "gauge" {
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Bad value", http.StatusBadRequest)
				return
			}
			store.SetGauge(metricName, val)
		}

		//Выполняем инкремент значения counter
		if metricType == "counter" {
			val, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Bad value", http.StatusBadRequest)
				return
			}
			store.AddCounter(metricName, val)
		}

	}
}

func GetCountMetric(store *repository.MemStore) http.HandlerFunc {
	return func(response http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		if metricType == "counter" {
			value, ok := store.GetCounter(metricName)
			if !ok {
				http.NotFound(response, r)
				return
			}
			if _, err := fmt.Fprint(response, strconv.FormatInt(value, 10)); err != nil {
				log.Printf("Failed to GetCountMetric: %v", err)
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
			return

		}
		if metricType == "gauge" {
			value, ok := store.GetGauge(metricName)
			if !ok {
				http.NotFound(response, r)
				return
			}
			if _, err := fmt.Fprint(response, strconv.FormatFloat(value, 'f', -1, 64)); err != nil {
				log.Printf("Failed to GetCountMetric: %v", err)
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
			return

		}
		http.NotFound(response, r)
	}
}

func AllMetrics(store *repository.MemStore) http.HandlerFunc {
	return func(response http.ResponseWriter, r *http.Request) {
		gauges := store.GetGauges()
		counters := store.GetCounters()
		t, err := template.New("templ").Parse(templateHTML)
		if err != nil {
			log.Printf("Failed to Allmetrics: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
		}
		if err := t.Execute(response, metrics{gauges, counters}); err != nil {
			log.Printf("Failed to Allmetrics: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func UpdateJSON(store *repository.MemStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Проверяем, что метод POST

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req models.Metrics
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to UpdateJson: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(buf, &req)
		logger.Log.Infow("Test", "error", err, "body", string(buf))
		if err != nil {
			log.Printf("Failed to UpdateJson: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Проверям, что введен правильный тип метрик
		if req.MType != "gauge" && req.MType != "counter" {
			http.Error(w, "Bad type of metric", http.StatusBadRequest)
			return
		}

		//Проверка, что имя метрики не пустое
		if req.ID == "" {
			http.NotFound(w, r)
			return
		}

		//Выполняем обновление значения gauge
		if req.MType == "gauge" {
			if req.Value == nil {
				http.Error(w, "Bad value", http.StatusBadRequest)
				return
			}
			store.SetGauge(req.ID, *req.Value)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
		}

		//Выполняем инкремент значения counter
		if req.MType == "counter" {
			if req.Delta == nil {
				http.Error(w, "Bad value", http.StatusBadRequest)
				return
			}
			store.AddCounter(req.ID, *req.Delta)
			//w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		}

	}
}

func ValueJSON(store *repository.MemStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Проверяем, что метод POST

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.Metrics
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to UpdateJson: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		logger.Log.Infow("Test", "body", string(buf))
		err = json.Unmarshal(buf, &req)

		if err != nil {
			log.Printf("Failed to UpdateJson: %v", err)
			//http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		//logger.Log.Infow("Test4")
		// Проверям, что введен правильный тип метрик
		if req.MType != "gauge" && req.MType != "counter" {
			http.Error(w, "Bad type of metric", http.StatusBadRequest)
			return
		}
		//logger.Log.Infow("Test5")

		//Проверка, что имя метрики не пустое
		if req.ID == "" {
			http.NotFound(w, r)
			return
		}
		//logger.Log.Infow("Test6")
		if req.MType == "counter" {
			value, ok := store.GetCounter(req.ID)
			if !ok {
				logger.Log.Infow("Нет метрики")
				http.NotFound(w, r)
				return
			}

			req.Delta = &value
			body, err := json.Marshal(req)

			if err != nil {
				log.Printf("Error marshalling json: %s\n", err)
				//return
			}

			/*if _, err := fmt.Fprint(w, body); err != nil {
				log.Printf("Failed to ValueJson: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}*/
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return

		}
		logger.Log.Infow("Test7")
		if req.MType == "gauge" {
			value, ok := store.GetGauge(req.ID)
			if !ok {
				http.NotFound(w, r)
				logger.Log.Infow("Нет метрики ")
				return
			}

			req.Value = &value
			body, err := json.Marshal(req)
			if err != nil {
				log.Printf("Error marshalling json: %s\n", err)
			}

			/*if _, err := fmt.Fprint(w, body); err != nil {
				log.Printf("Failed to ValueJson: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}*/
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return

		}

	}
}
