package handler

import (
	"context"
	"database/sql"
	"errors"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// updateOneMetric записывыает изменения одной метрики в базу
func updateOneMetric(req models.Metrics, db *sql.DB, r *http.Request) (errorTxt string, code int) {
	// Проверям, что введен правильный тип метрик
	errorTxt = ""
	code = 0
	if req.MType != "gauge" && req.MType != "counter" {
		//HTTPError(w, "Bad type of metric", http.StatusBadRequest)
		errorTxt = "Bad type of metric"
		code = http.StatusBadRequest
		return errorTxt, code
	}

	//Проверка, что имя метрики не пустое
	if req.ID == "" {
		//http.NotFound(w, r)
		errorTxt = "NotFound"
		code = http.StatusNotFound
		return errorTxt, code
	}

	//Для типа Counter получаем предыдущее значение для суммирования
	logger.Log.Infow("До oldmetric", "id", req.ID)

	if req.MType == "counter" {
		var oldMetric3 models.Metrics
		if req.Delta == nil {
			//HTTPError(w, "Bad value", http.StatusBadRequest)
			errorTxt = "Bad value"
			code = http.StatusBadRequest
			return errorTxt, code
		}
		query := `
					SELECT name, value, delta, type 
					FROM metrics
					WHERE name = $1 AND type = $2
					`

		logger.Log.Infow("До проверки", "id", req.ID)
		//Проверяем есть ли в базе метрика
		oldMetric3, err := requestSelectDB(r.Context(), db, req, query)
		if err != nil {
			logger.Log.Infow("Ошибка чтения метрик")
			//return
		}
		logger.Log.Infow("После запроса")
		//if len(oldName) > 0 {
		if len(oldMetric3.ID) > 0 {
			logger.Log.Infow("строка не пустая")
			newDelta := *req.Delta + *oldMetric3.Delta
			req.Delta = &newDelta
		} else {
			logger.Log.Infow("строка пустая")
		}
	}
	if req.MType == "gauge" {
		if req.Value == nil {
			//HTTPError(w, "Bad value", http.StatusBadRequest)
			errorTxt = "Bad value"
			code = http.StatusBadRequest
			return errorTxt, code
		}
	}

	// Запись в базу новых метрик
	err := requestDB(r.Context(), db, req, queryUpdate)

	if err != nil {
		logger.Log.Error("failed to add metric", zap.Error(err))
		return errorTxt, code
	}
	return errorTxt, code
}

// readOneMetric считывает одну метрику из базы
func readOneMetric(req models.Metrics, db *sql.DB, r *http.Request) (reqDB models.Metrics, code int, errorTxt string) {
	// Проверям, что введен правильный тип метрик
	if req.MType != "gauge" && req.MType != "counter" {
		//HTTPError(w, "Bad type of metric", http.StatusBadRequest)
		errorTxt = "Bad type of metric"
		code = http.StatusBadRequest
		return reqDB, code, errorTxt
	}

	//Проверка, что имя метрики не пустое
	if req.ID == "" {
		//http.NotFound(w, r)
		//HTTPError(w, "{}", http.StatusNotFound)
		errorTxt = ""
		code = http.StatusNotFound
		return reqDB, code, errorTxt
	}

	query := `
					SELECT name, value, delta, type 
					FROM metrics
					WHERE name = $1 AND type = $2
					`
	reqDB, err := requestSelectDB(r.Context(), db, req, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Log.Debug("metric not found", zap.String("name", reqDB.ID))
			//HTTPError(w, "{}", http.StatusNotFound)
			errorTxt = ""
			code = http.StatusNotFound
			return reqDB, code, errorTxt
		}
		logger.Log.Infow("failed to get metric", zap.Error(err))
		//HTTPError(w, "{}", http.StatusNotFound)
		errorTxt = ""
		code = http.StatusNotFound
		return reqDB, code, errorTxt
	}
	return reqDB, code, errorTxt

}

// readAllMetrics считывает все метрики из базы
func readAllMetrics(db *sql.DB, r *http.Request) (allMetrics metrics, code int, errorTxt string) {
	query := `
		SELECT name, value, delta, type
		FROM metrics;
		`
	//gauges := make(map[string]float64)
	//counters := make(map[string]int64)

	//rows, err := db.Query(query)
	rows, err := requestSelectAllDB(r.Context(), db, query)
	if rows.Err() != nil {
		logger.Log.Errorw("<UNK> <UNK>", "query", query)
		return allMetrics, code, errorTxt
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Log.Debug("metric not found")
			return allMetrics, code, errorTxt
		}
		logger.Log.Infow("failed to list metrics", zap.Error(err))
		return allMetrics, code, errorTxt
	}
	//var metricsDB []*models.Metrics
	logger.Log.Infow("До rows.Next")
	for rows.Next() {
		var metric models.Metrics
		err = rows.Scan(&metric.ID, &metric.Value, &metric.Delta, &metric.MType)
		if err != nil {
			logger.Log.Error("failed to scan metric", zap.Error(err))
			continue
			//return allMetrics, code, errorTxt
		}
		//logger.Log.Infow("Перед присвоением метрик gauge", "gauge", *metric.Value, "id", metric.ID)
		if metric.MType == "gauge" {
			logger.Log.Infow("Перед присвоением метрик gauge", "gauge", *metric.Value, "id", metric.ID)
			if metric.Value != nil {
				allMetrics.Gauge[metric.ID] = *metric.Value
			}
		}
		//logger.Log.Infow("Перед присвоением метрик counter", "counter", *metric.Delta)
		if metric.MType == "counter" {
			logger.Log.Infow("Перед присвоением метрик counter", "counter", *metric.Delta, "id", metric.ID)
			if metric.Delta != nil {
				allMetrics.Counter[metric.ID] = *metric.Delta
			}
		}
		logger.Log.Infow("После counter")
		return allMetrics, code, errorTxt
	}
	return allMetrics, code, errorTxt
}

// updateManyMetrics считывает все метрики из базы
func updateManyMetrics(reqs []models.Metrics, db *sql.DB, r *http.Request) (code int, errorTxt string) {
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
			//HTTPError(w, "internal server error", http.StatusInternalServerError)
			errorTxt = "internal server error"
			code = http.StatusInternalServerError
			return code, errorTxt
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
	return
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
		case <-ctx.Done():
			return err
		}
	}
	return err
}

func requestSelectDB(ctx context.Context, db *sql.DB, req models.Metrics, query string) (oldMetric models.Metrics, err error) {

	timer := time.NewTimer(time.Duration(0) * time.Second)
	defer timer.Stop()
	var pgErr *pgconn.PgError
	for i := 1; i <= 5; i += 2 {

		err = db.QueryRow(query, req.ID, req.MType).Scan(
			//&metric, &name,
			&oldMetric.ID, &oldMetric.Value, &oldMetric.Delta, &oldMetric.MType,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Log.Infow("<UNK> <UNK>", "id", req.ID)
			}
		}

		if !(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)) {
			//return metric, name, err
			return oldMetric, err
		}
		timer.Reset(time.Duration(i) * time.Second)
		select {
		case <-timer.C:
			logger.Log.Infow("Ошибка при подключении к базе")
		case <-ctx.Done():
			//return nil, "", err
			return oldMetric, err
		}
	}
	//return nil, "", err
	return oldMetric, err
}

func requestSelectAllDB(ctx context.Context, db *sql.DB, query string) (rows *sql.Rows, err error) {

	timer := time.NewTimer(time.Duration(0) * time.Second)
	defer timer.Stop()
	var pgErr *pgconn.PgError
	for i := 1; i <= 5; i += 2 {

		rows, err := db.Query(query)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Log.Infow("Нет строк")
			}
		}

		if !(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)) {
			//return metric, name, err
			return rows, err
		}
		timer.Reset(time.Duration(i) * time.Second)
		select {
		case <-timer.C:
			logger.Log.Infow("Ошибка при подключении к базе")
		case <-ctx.Done():
			//return nil, "", err
			return rows, err
		}
	}
	//return nil, "", err
	return rows, err
}

func requestCommitDB(ctx context.Context, db *sql.DB, tx *sql.Tx) (err error) {

	timer := time.NewTimer(time.Duration(0) * time.Second)
	defer timer.Stop()
	var pgErr *pgconn.PgError
	for i := 1; i <= 5; i += 2 {

		err = tx.Commit()
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Log.Infow("Нет строк")
			}
		}

		if !(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)) {
			//return metric, name, err
			return err
		}
		timer.Reset(time.Duration(i) * time.Second)
		select {
		case <-timer.C:
			logger.Log.Infow("Ошибка при подключении к базе")
		case <-ctx.Done():
			//return nil, "", err
			return err
		}
	}
	//return nil, "", err
	return err
}
