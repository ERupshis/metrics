package main

import (
	"net/http"
	"time"

	"github.com/erupshis/metrics/internal/agent/ticker"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/controllers"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemanager"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Parse()

	log := logger.CreateLogger(cfg.LogLevel)
	defer log.Sync()

	storageManager := storagemanager.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(&storageManager)

	baseController := controllers.CreateBase(cfg, log, storage)

	router := chi.NewRouter()
	router.Mount("/", baseController.Route())

	//Schedule data saving in file with storeInterval
	storeTicker := scheduleDataStoringInFile(&cfg, storage, &log)
	defer storeTicker.Stop()

	log.Info("Server started with Host setting: %s", cfg.Host)
	if err := http.ListenAndServe(cfg.Host, router); err != nil {
		panic(err)
	}
}

func scheduleDataStoringInFile(cfg *config.Config, storage memstorage.MemStorage, log *logger.BaseLogger) *time.Ticker {
	var interval int64 = 1
	if cfg.StoreInterval > 1 {
		interval = cfg.StoreInterval
	}

	(*log).Info("[main::scheduleDataStoringInFile] init saving in file with interval: %d", cfg.StoreInterval)
	storeTicker := ticker.CreateWithSecondsInterval(interval)
	go ticker.Run(storeTicker, func() { storage.SaveData() })
	return storeTicker
}
