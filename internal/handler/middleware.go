package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/config"
	"io"
	"net/http"
)

func SetContentType(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, r)
	})
}

// CheckSign проверка подписи запроса
func CheckSign(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hashHeader := r.Header.Get("HashSHA256"); hashHeader != "" && config.Params.HashKey != "" {
			hash, err := hex.DecodeString(hashHeader)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Читаем тело запроса
			rawBody, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// Ставим тело снова, чтобы его можно было прочитать снова.
			r.Body = io.NopCloser(bytes.NewBuffer(rawBody))

			harsher := hmac.New(sha256.New, []byte(config.Params.HashKey))
			harsher.Write(rawBody)
			hashSum := harsher.Sum(nil)
			if !hmac.Equal(hash, hashSum) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
