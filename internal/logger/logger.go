package logger

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type logger struct {
	zap *zap.Logger
}

func CreateZapLogger(level string) (*logger, error) {
	cfg, err := initConfig(level)
	if err != nil {
		return nil, err
	}

	log, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &logger{zap: log}, nil
}

func (l *logger) Info(msg string, fields ...interface{}) {
	l.zap.Info(fmt.Sprintf(msg, fields...))
}

func initConfig(level string) (zap.Config, error) {
	cfg := zap.NewProductionConfig()

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		emptyConfig := zap.Config{}
		return emptyConfig, err
	}
	cfg.Level = lvl

	return cfg, nil
}

func (l *logger) Sync() {
	err := l.zap.Sync()
	if err != nil {
		panic(err)
	}
}

func (l *logger) LogHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		loggingWriter := createResponseWriter(w)
		h.ServeHTTP(loggingWriter, r)
		duration := time.Since(start)

		l.zap.Info("new incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", loggingWriter.getResponseData().status),
			zap.String("content-type", loggingWriter.Header().Get("Content-Type")),
			zap.String("content-encoding", loggingWriter.Header().Get("Content-Encoding")),
			zap.Duration("duration", duration),
			zap.Int("size", loggingWriter.getResponseData().size),
		)
	})
}
