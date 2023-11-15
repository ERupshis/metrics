package logger

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var (
	_ BaseLogger = (*zapLogger)(nil)
)

// zapLogger BaseLogger implementation based on Zap.
type zapLogger struct {
	zap *zap.Logger
}

// CreateZapLogger returns base logger
func CreateZapLogger(level string) (BaseLogger, error) {
	cfg, err := initConfig(level)
	if err != nil {
		return nil, err
	}

	logTmp, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("create zap logger %w", err)
	}

	return &zapLogger{zap: logTmp}, nil
}

func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.zap.Info(fmt.Sprintf(msg, fields...))
}

// initConfig config generator.
func initConfig(level string) (zap.Config, error) {
	cfg := zap.NewProductionConfig()

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		emptyConfig := zap.Config{}
		return emptyConfig, fmt.Errorf("init zap logger config: %w", err)
	}
	cfg.Level = lvl
	cfg.DisableCaller = true

	return cfg, nil
}

func (l *zapLogger) Sync() {
	err := l.zap.Sync()
	if err != nil {
		log.Printf("[loggerZap:Sync] error occured: %v\n", err)
	}
}

func (l *zapLogger) LogHandler(h http.Handler) http.Handler {
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
			zap.String("HashSHA256", loggingWriter.Header().Get("HashSHA256")),
			zap.Duration("duration", duration),
			zap.Int("size", loggingWriter.getResponseData().size),
		)
	})
}
