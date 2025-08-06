// Package logging пакет с описанием логирования
package logging

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var sugar zap.SugaredLogger

// ResponseData содержит данные о gRPC ответе
type GrpcResponseData struct {
	Status     string
	StatusCode int
	Duration   time.Duration
	Method     string
}

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

// loggingUnaryServerInterceptor возвращает UnaryServerInterceptor для логирования
func LoggingUnaryInterceptor(logger zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Вызываем обработчик
		res, err := handler(ctx, req)

		// Получаем статус ответа
		st, _ := status.FromError(err)
		statusCode := st.Code()
		statusMsg := st.Message()

		// Формируем данные для логирования
		responseData := &GrpcResponseData{
			Status:     statusMsg,
			StatusCode: int(statusCode),
			Duration:   time.Since(start),
			Method:     info.FullMethod,
		}

		// Логируем информацию
		logger.Infoln(
			"\n",
			"-----GRPC REQUEST-----\n",
			"Method:", responseData.Method, "\n",
			"Status:", responseData.Status, "\n",
			"Duration:", responseData.Duration, "\n",
			"Status code:", responseData.StatusCode, "\n",
		)

		return res, err
	}
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
