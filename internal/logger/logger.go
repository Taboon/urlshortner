package logger

import (
	"fmt"
	"github.com/Taboon/urlshortner/internal/config"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

type Logger struct {
	*zap.Logger
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func Initialize(conf config.Config) (*zap.Logger, error) {
	var logger = &zap.Logger{}
	lvl, err := zap.ParseAtomicLevel(conf.LogLevel)
	if err != nil {
		return logger, err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return logger, err
	}
	logger = zl
	return logger, nil
}

func (l *Logger) RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		uri := r.RequestURI
		method := r.Method
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		l.Info("request", //nolint:typecheck
			zap.String("uri", uri),
			zap.String("method", method),
			zap.String("duration", strconv.FormatInt(int64(duration), 10)),
			zap.String("response status", strconv.Itoa(responseData.status)),
			zap.String("response size", strconv.Itoa(responseData.size)),
		)
		fmt.Println(r)
	}
	return logFn
}
