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
		/*
			var (
				counters = make(map[string]int64)
			)

			//Начало транзакции
			tx, err := db.Begin()
			if err != nil {
				logger.Log.Infow("Ошибка начала транзакции", "err", err)
				return
			}
			defer tx.Rollback()

			for _, req := range reqs {
				if req.ID == "" {
					HTTPError(w, "internal server error", http.StatusInternalServerError)
					return
				}
				if req.MType == "counter" {
					//Для типа Counter получаем предыдущее значение для суммирования
					logger.Log.Infow("Counter До oldmetric", "id", req.ID)
					var oldMetric2 models.Metrics
					var newDelta int64

					query := `
						SELECT name, value, delta, type
						FROM metrics
						WHERE name = $1 AND type = $2
						`
					logger.Log.Infow("До проверки", "id", req.ID)
					oldMetric2, err = requestSelectDB(r.Context(), db, req, query)
					if err != nil {
						if err == sql.ErrNoRows {
							logger.Log.Infow("<UNK> <UNK>", "id", req.ID)
						}
					}
					logger.Log.Infow("После запроса")

					if _, exists := counters[req.ID]; exists {
						logger.Log.Infow("Метрика есть", "newDelta = ", newDelta)
						newDelta = counters[req.ID] + *req.Delta
						logger.Log.Infow("Добавили к существующей", "добавили", newDelta, "counters", counters[req.ID])

					} else {
						if len(oldMetric2.ID) > 0 {
							newDelta = *req.Delta + *oldMetric2.Delta
							logger.Log.Infow("строка не пустая", "newDelta", newDelta, "oldMetric", oldMetric2.Delta)
						} else {
							newDelta = *req.Delta
							logger.Log.Infow("строка пустая", "newDelta", newDelta)
						}
						logger.Log.Infow("Метрика отсутствует")
						logger.Log.Infow("Новая NewDelta", "добавили", newDelta, "id", req.ID)
					}
					req.Delta = &newDelta
					counters[req.ID] = newDelta
				}

				_, err := tx.Exec(queryUpdate, req.ID, req.Value, req.Delta, req.MType)
				if err != nil {
					logger.Log.Infow("<UNK> <UNK> <UNK>", "err", err)
					return
				}

			}
			err = requestCommitDB(r.Context(), db, tx)
			if err != nil {
				logger.Log.Infow("<UNK> <UNK> <UNK>", "err", err)
				return
			}
		*/
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
