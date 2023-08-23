package main

import (
	"net/http"

	"github.com/erupshis/metrics/internal/agent/ticker"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/controllers"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Parse()

	log := logger.CreateLogger(cfg.LogLevel)
	defer log.Sync()

	baseController := controllers.CreateBase(cfg, log)

	router := chi.NewRouter()
	router.Mount("/", baseController.Route())

	storeTicker := ticker.CreateWithSecondsInterval( /*cfg.StoreInterval*/ 10)
	defer storeTicker.Stop()
	go ticker.Run(storeTicker, func() { baseController.SaveMetricsInFile() })

	log.Info("Server started with Host setting: %s", cfg.Host)
	if err := http.ListenAndServe(cfg.Host, router); err != nil {
		panic(err)
	}
}
