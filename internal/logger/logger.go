package logger

import (
	"net/http"

	"go.uber.org/zap"
)

type InfoLogger struct {
	zap *zap.Logger
}

func Create(level string) (*InfoLogger, error) {
	cfg, err := initConfig(level)
	if err != nil {
		return nil, err
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &InfoLogger{zap: logger}, nil
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

func (il *InfoLogger) Sync() {
	err := il.zap.Sync()
	if err != nil {
		panic(err)
	}
}

func (il *InfoLogger) RequestLogger(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		il.zap.Debug("got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		h(w, r)
	})
}
