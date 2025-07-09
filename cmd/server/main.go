package main

import (
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type metrics struct {
	Gauge   map[string]float64
	Counter map[string]int64
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

func updatePage(w http.ResponseWriter, r *http.Request) {
	//Проверяем, что метод POST
	if r.Method != http.MethodPost {
		//fmt.Fprint(w, "method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		//w.Write([]byte("method not allowed"))
		//w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	// Проверям, что введен правильный тип метрик
	if metricType != "gauge" && metricType != "counter" {
		//fmt.Fprint(w, "Bad request")
		//w.WriteHeader(http.StatusBadRequest)
		//w.WriteHeader(http.StatusBadRequest)
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
			//fmt.Fprint(w, "Bad request")
			//w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Bad value", http.StatusBadRequest)
			return
		}
		repository.Store.SetGauge(metricName, val)
	}

	//Выполняем инкремент значения counter
	if metricType == "counter" {
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Bad value", http.StatusBadRequest)
			return
		}
		repository.Store.AddCounter(metricName, val)
	}

}

func getCountMetric(response http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	if metricType == "counter" {
		value, ok := repository.Store.GetCounter(metricName)
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
		value, ok := repository.Store.GetGauge(metricName)
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

func allMetrics(response http.ResponseWriter, r *http.Request) {
	//repository.Store.SetGauge("test", 5)
	gauges := repository.Store.GetGauges()
	counters := repository.Store.GetCounters()
	t, err := template.New("templ").Parse(templateHTML)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
	if err := t.Execute(response, metrics{gauges, counters}); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

}

func main() {

	// обрабатываем аргументы командной строки
	config, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	repository.Store = repository.New()
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", updatePage)
	r.Get("/value/{type}/{name}", getCountMetric)
	r.Get("/", allMetrics)

	if err := runServer(r, config.AddrServer); err != nil {
		panic(err)
	}

	/*err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}*/
}

func runServer(router chi.Router, addr string) error {
	fmt.Println("Starting server", addr)
	return http.ListenAndServe(addr, router)
}
