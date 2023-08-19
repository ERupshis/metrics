package logger

import (
	"bytes"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type RequestLogger struct {
	zap *zap.Logger
}

func CreateRequest(level string) (*RequestLogger, error) {
	cfg, err := initConfig(level)
	if err != nil {
		return nil, err
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &RequestLogger{zap: logger}, nil
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

func (il *RequestLogger) Sync() {
	err := il.zap.Sync()
	if err != nil {
		panic(err)
	}
}

func (il *RequestLogger) Log(h http.HandlerFunc) http.HandlerFunc {
	logWrap := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		loggingWriter := createResponseWriter(w)
		h.ServeHTTP(loggingWriter, r)
		duration := time.Since(start)

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r.Body)
		defer r.Body.Close()

		il.zap.Info("new incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("body", buf.String()),
			zap.Int("status", loggingWriter.getResponseData().status),
			zap.Duration("duration", duration),
			zap.Int("size", loggingWriter.getResponseData().size),
		)
	}

	return logWrap
}
