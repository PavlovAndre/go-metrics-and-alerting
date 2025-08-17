package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
)

const queryUpdate = `
					INSERT INTO metrics (name, value, delta, type)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (name) DO UPDATE
					SET value = EXCLUDED.value, delta = EXCLUDED.delta, type = EXCLUDED.type;
					`

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

		// Проверям, что введен правильный тип метрик
		if req.MType != "gauge" && req.MType != "counter" {
			HTTPError(w, "Bad type of metric", http.StatusBadRequest)
			return
		}

		//Проверка, что имя метрики не пустое
		if req.ID == "" {
			http.NotFound(w, r)
			return
		}

		//Для типа Counter получаем предыдущее значение для суммирования
		logger.Log.Infow("До oldmetric", "id", req.ID)
		var oldMetric *int64
		var oldName string
		if req.MType == "counter" {
			if req.Delta == nil {
				HTTPError(w, "Bad value", http.StatusBadRequest)
				return
			}
			query := `
					SELECT delta, name
					FROM metrics
					WHERE name = $1
					`

			logger.Log.Infow("До проверки", "id", req.ID)
			err = db.QueryRow(query, req.ID).Scan(
				&oldMetric, &oldName,
			)
			if err != nil {
				if err == sql.ErrNoRows {
					logger.Log.Infow("<UNK> <UNK>", "id", req.ID)
				}
			}
			logger.Log.Infow("После запроса")
			if len(oldName) > 0 {
				logger.Log.Infow("строка не пустая")
				newDelta := *req.Delta + *oldMetric
				req.Delta = &newDelta
			} else {
				logger.Log.Infow("строка пустая")
			}
		}
		if req.MType == "gauge" {
			if req.Value == nil {
				HTTPError(w, "Bad value", http.StatusBadRequest)
				return
			}
		}
		// Запись в базу новых метрик
		query := `
					INSERT INTO metrics (name, value, delta, type)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (name) DO UPDATE
					SET value = EXCLUDED.value, delta = EXCLUDED.delta, type = EXCLUDED.type;
					`
		//timer := time.NewTimer(time.Duration(0) * time.Second)
		//defer timer.Stop()

		err = requestDB(r.Context(), db, req, query)

		if err != nil {
			logger.Log.Error("failed to add metric", zap.Error(err))
			return
		}
		w.Write([]byte("{}"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		logger.Log.Debug("metric added successfully", zap.String("name", req.ID))
		//return
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
		// Проверям, что введен правильный тип метрик
		if req.MType != "gauge" && req.MType != "counter" {
			HTTPError(w, "Bad type of metric", http.StatusBadRequest)
			return
		}

		//Проверка, что имя метрики не пустое
		if req.ID == "" {
			//http.NotFound(w, r)
			HTTPError(w, "{}", http.StatusNotFound)
			return
		}

		query := `
					SELECT name, value, delta, type 
					FROM metrics
					WHERE name = $1 AND type = $2
					`

		err = db.QueryRow(query, req.ID, req.MType).Scan(
			&req.ID, &req.Value, &req.Delta, &req.MType,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				logger.Log.Debug("metric not found", zap.String("name", req.ID))
				HTTPError(w, "{}", http.StatusNotFound)
				return
			}
			logger.Log.Infow("failed to get metric", zap.Error(err))
			HTTPError(w, "{}", http.StatusNotFound)
			return
		}

		logger.Log.Infow("До каунтер")
		if req.MType == "counter" {
			logger.Log.Infow("Тип каутнер")
			body, err := json.Marshal(req)

			if err != nil {
				logger.Log.Infow("Error marshalling json: %s\n", err)
				HTTPError(w, "{}", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return

		}
		if req.MType == "gauge" {
			body, err := json.Marshal(req)
			if err != nil {
				log.Printf("Error marshalling json: %s\n", err)
				HTTPError(w, "{}", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return

		}

	}
}

func AllDB(db *sql.DB) http.HandlerFunc {
	return func(response http.ResponseWriter, r *http.Request) {
		logger.Log.Infow("Start AllDB")
		query := `
		SELECT name, value, delta, type
		FROM metrics;
		`
		gauges := make(map[string]float64)
		counters := make(map[string]int64)

		rows, err := db.Query(query)
		if rows.Err() != nil {
			logger.Log.Errorw("<UNK> <UNK>", "query", query)
			return
		}
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				logger.Log.Debug("metric not found")
				return
			}
			logger.Log.Infow("failed to list metrics", zap.Error(err))
			return
		}
		//var metricsDB []*models.Metrics
		logger.Log.Infow("До rows.Next")
		for rows.Next() {
			var metric models.Metrics
			err = rows.Scan(&metric.ID, &metric.Value, &metric.Delta, &metric.MType)
			if err != nil {
				logger.Log.Error("failed to scan metric", zap.Error(err))
				continue
			}
			//logger.Log.Infow("Перед присвоением метрик gauge", "gauge", *metric.Value, "id", metric.ID)
			if metric.MType == "gauge" {
				logger.Log.Infow("Перед присвоением метрик gauge", "gauge", *metric.Value, "id", metric.ID)
				if metric.Value != nil {
					gauges[metric.ID] = *metric.Value
				}
			}
			//logger.Log.Infow("Перед присвоением метрик counter", "counter", *metric.Delta)
			if metric.MType == "counter" {
				logger.Log.Infow("Перед присвоением метрик counter", "counter", *metric.Delta, "id", metric.ID)
				if metric.Delta != nil {
					counters[metric.ID] = *metric.Delta
				}
			}
			logger.Log.Infow("После counter")

		}
		logger.Log.Infow("<UNK> <UNK>", "gauges", gauges, "counters", counters)
		t, err := template.New("templ").Parse(TemplateHTML)
		if err != nil {
			log.Printf("Failed to Allmetrics: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
		}
		if err := t.Execute(response, metrics{gauges, counters}); err != nil {
			log.Printf("Failed to Allmetrics: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "application/json")
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

		var (
			//gauges   = make(map[string]float64)
			counters = make(map[string]int64)
		)
		logger.Log.Infow("Значение мапы counters", "counters", counters)
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
				var oldMetric *int64
				var oldName string
				var newDelta int64
				query := `
					SELECT delta, name
					FROM metrics
					WHERE name = $1
					`

				logger.Log.Infow("До проверки", "id", req.ID)
				err = db.QueryRow(query, req.ID).Scan(
					&oldMetric, &oldName,
				)
				if err != nil {
					if err == sql.ErrNoRows {
						logger.Log.Infow("<UNK> <UNK>", "id", req.ID)
					}
				}
				logger.Log.Infow("После запроса")

				if _, exists := counters[req.ID]; exists {
					logger.Log.Infow("Метрика есть", "newDelta = ", newDelta)
					//newDelta = newDelta + counters[req.ID]
					newDelta = counters[req.ID] + *req.Delta
					logger.Log.Infow("Добавили к существующей", "добавили", newDelta, "counters", counters[req.ID])

				} else {
					if len(oldName) > 0 {
						newDelta = *req.Delta + *oldMetric
						logger.Log.Infow("строка не пустая", "newDelta", newDelta, "oldMetric", oldMetric)
						//req.Delta = &newDelta
					} else {
						newDelta = *req.Delta
						logger.Log.Infow("строка пустая", "newDelta", newDelta)
					}
					logger.Log.Infow("Метрика отсутствует")
					//newDelta = *req.Delta
					logger.Log.Infow("Новая NewDelta", "добавили", newDelta, "id", req.ID)
				}
				req.Delta = &newDelta
				counters[req.ID] = newDelta
				//logger.Log.Infow("Значение мапы2 counters", "counters", counters)
			}

			_, err := tx.Exec(queryUpdate, req.ID, req.Value, req.Delta, req.MType)
			//logger.Log.Infow("tx.Exec", "tx.Exec", arts)
			if err != nil {
				logger.Log.Infow("<UNK> <UNK> <UNK>", "err", err)
				return
			}

		}
		err = tx.Commit()
		if err != nil {
			logger.Log.Infow("<UNK> <UNK> <UNK>", "err", err)
			return
		}
		logger.Log.Infow("Метрики добавлены")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func requestDB(ctx context.Context, db *sql.DB, req models.Metrics, query string) (err error) {
	timer := time.NewTimer(time.Duration(0) * time.Second)
	defer timer.Stop()
	var pgErr *pgconn.PgError
	for i := 1; i <= 5; i += 2 {

		_, err = db.Exec(query, req.ID, req.Value, req.Delta, req.MType)
		if err == nil {
			logger.Log.Infow("Подключились к базе без ошибок")
			return nil
		}

		if !(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)) {
			return err
		}
		timer.Reset(time.Duration(i) * time.Second)
		select {
		case <-timer.C:
			logger.Log.Infow("Ошибка при подключении к базе")
			//timer.Reset(time.Duration(i) * time.Second)
		case <-ctx.Done():
			return err
		}
	}
	return err
}
