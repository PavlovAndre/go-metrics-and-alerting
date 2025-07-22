package logger

import (
	"fmt"
	"net/http"
	"time"
)

// LogRequest  мидлеваре, которое регистрирует данные запроса
// Функция регистрирует метод, путь и продолжительность каждого запроса
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			Log.Infow("Incoming request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start))
		}()
		next.ServeHTTP(w, r)
	})
}

// берём структуру для хранения сведений об ответе
type responseData struct {
	status int
	size   int
}

// добавляем реализацию http.ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	data *responseData
}

// Write реализует метод http.ResponseWriter.Write интерфейса http.ResponseWriter
// Заполняет размер передаваемых данных тела
func (r *responseWriter) Write(body []byte) (int, error) {
	size, err := r.ResponseWriter.Write(body)
	r.data.size += size
	return size, err
}

// WriteHeader реализует метод http.ResponseWriter.WriteHeader интерфейса http.ResponseWriter
// Сохраняет статус ответа
func (r *responseWriter) WriteHeader(status int) {
	r.data.status = status
	r.ResponseWriter.WriteHeader(status)
}

// LogResponse мидлеваре, которое регистрирует данные ответа
// Функция регистрирует размер тела и статус ответа
func LogResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, r *http.Request) {
		newWriter := &responseWriter{ResponseWriter: response, data: new(responseData)}
		defer func() {
			Log.Infow("Sent response",
				"status", newWriter.data.status,
				"size", fmt.Sprintf("%d B", newWriter.data.size))
		}()
		next.ServeHTTP(newWriter, r)

	})
}
