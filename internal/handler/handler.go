package handler

import (
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"html/template"
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
				http.Error(response, err.Error(), http.StatusInternalServerError)
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
			if _, err := fmt.Fprint(response, strconv.FormatFloat(value, 'f', 3, 64)); err != nil {
				http.Error(response, err.Error(), http.StatusInternalServerError)
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
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		if err := t.Execute(response, metrics{gauges, counters}); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}
