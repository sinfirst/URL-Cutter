// Package logging пакет с описанием логирования
package logging

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

// ResponseData структура для данных из запроса
type ResponseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *ResponseData
}

// Write запись данных
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader запись хеда из запроса
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// ResponseWriter интерфейс для записи данных
type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

// WithLogging прослойка для логирования запросов
func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &ResponseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		sugar.Infoln(
			"\n",
			"-----REQUEST-----\n",
			"URI:", r.RequestURI, "\n",
			"Method:", r.Method, "\n",
			"Duration:", duration, "\n",
			"-----RESPONSE-----\n",
			"Status:", responseData.status, "\n",
			"Size:", responseData.size, "\n",
		)
	})
}

// NewLogger конструктор для структуры
func NewLogger() zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	sugar = *logger.Sugar()

	return sugar
}
