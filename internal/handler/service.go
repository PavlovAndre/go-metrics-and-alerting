package handler

import (
	"database/sql"
	"errors"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"net/http"
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
