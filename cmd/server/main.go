package main

import (
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"net/http"
	"strconv"
	"strings"
)

func updatePage(w http.ResponseWriter, r *http.Request) {
	//Проверяем, что метод POST
	if r.Method != http.MethodPost {
		//fmt.Fprint(w, "method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		//w.Write([]byte("method not allowed"))
		//w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//Получаем входящий URL и разбиваем по переменным
	urlRequest := r.URL.Path
	sliceUrl := strings.Split(urlRequest, "/")
	fmt.Println(sliceUrl)
	if len(sliceUrl) != 5 {
		http.NotFound(w, r)
		return
	}
	metricType := sliceUrl[2]
	metricName := sliceUrl[3]
	metricValue := sliceUrl[4]

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
		repository.Store.SetCounter(metricName, val)
	}

	return
}

func main() {
	repository.Store = repository.New()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", updatePage)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
