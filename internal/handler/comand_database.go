package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"go.uber.org/zap"
	"html/template"
	"io"
	"log"
	"net/http"
)

func UpdateDB(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Infow("Запущена функция UpdateDB")

		//Проверяем, что метод POST
		if r.Method != http.MethodPost {
			HTTPError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req models.Metrics
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to UpdateJson: %v", err)
			HTTPError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(buf, &req)
		logger.Log.Infow("Лог UpdateDB", "error", err, "body", string(buf))
		if err != nil {
			log.Printf("Failed to UpdateJson: %v", err)
			HTTPError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		//Отправляем запрос в базу
		errorText, code := updateOneMetric(req, db, r)

		if code != 0 {
			HTTPError(w, errorText, code)
			return
		}

		w.Write([]byte("{}"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		logger.Log.Debug("metric added successfully", zap.String("name", req.ID))
	}
}

func ValueDB(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Infow("Запущена функция ValueDB")
		w.Header().Set("Content-Type", "application/json")

		//Проверяем, что метод POST
		if r.Method != http.MethodPost {
			HTTPError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.Metrics
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Infow("Failed read buf.", "ошибка:", err, "buf:", string(buf))
			HTTPError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		logger.Log.Infow("Лог ValueDB", "body", string(buf))
		err = json.Unmarshal(buf, &req)

		if err != nil {
			logger.Log.Infow("Failed to UpdateJson.", "ошибка:", err)
			HTTPError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		//Отправляем запрос в базу
		req, code, errorText := readOneMetric(req, db, r)

		if code != 0 {
			HTTPError(w, errorText, code)
			return
		}

		logger.Log.Infow("До каунтер")

		body, err := json.Marshal(req)
		if err != nil {
			log.Printf("Error marshalling json: %s\n", err)
			HTTPError(w, "{}", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func AllDB(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Infow("Start AllDB")
		metricDB, code, errorText := readAllMetrics(db, r)

		if code != 0 {
			HTTPError(w, errorText, code)
			return
		}
		logger.Log.Infow("<UNK> <UNK>", "gauges", metricDB.Gauge, "counters", metricDB.Counter)
		t, err := template.New("templ").Parse(TemplateHTML)
		if err != nil {
			log.Printf("Failed to Allmetrics: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if err := t.Execute(w, metricDB); err != nil {
			log.Printf("Failed to Allmetrics: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
	}
}

func UpdatesDB(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Infow("Запущена функция UpdatesDB")
		//Проверяем, что метод POST
		if r.Method != http.MethodPost {
			HTTPError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		//Читаем тело запроса
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to UpdateJson: %v", err)
			HTTPError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Парсим тело в структуру запроса
		var reqs []models.Metrics
		err = json.Unmarshal(buf, &reqs)
		logger.Log.Infow("Лог UpdatesDB", "error", err, "body", string(buf))
		if err != nil {
			log.Printf("Failed to UpdateJson: %v", err)
			HTTPError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		//Отправляем запрос в базу
		code, errorText := updateManyMetrics(reqs, db, r)

		if code != 0 {
			HTTPError(w, errorText, code)
			return
		}
		logger.Log.Infow("Метрики добавлены")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
