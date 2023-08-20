package main

import (
	"net/http"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/controllers"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Parse()

	log := createLogger(cfg.LogLevel)
	defer log.Sync()

	baseController := controllers.CreateBase(cfg, log)

	router := chi.NewRouter()
	router.Mount("/", baseController.Route())

	log.Info("Server started with Host setting: %s", cfg.Host)
	if err := http.ListenAndServe(cfg.Host, router); err != nil {
		panic(err)
	}
}

func createLogger(level string) logger.BaseLogger {
	log, err := logger.CreateZapLogger(level)
	if err != nil {
		panic(err)
	}

	return log
}
