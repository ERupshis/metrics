package main

import (
	"context"
	"net/http"
	"time"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/controllers"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
	"github.com/erupshis/metrics/internal/ticker"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Parse()

	log := logger.CreateLogger(cfg.LogLevel)
	defer log.Sync()

	storageManager, err := createStorageManager(&cfg, log)
	if err != nil {
		log.Info("[main] failed to create connection to database: %s", cfg.DataBaseDSN)
	}
	if storageManager != nil {
		defer storageManager.Close()
	}
	storage := memstorage.Create(storageManager)

	baseController := controllers.CreateBase(cfg, log, storage)

	router := chi.NewRouter()
	router.Mount("/", baseController.Route())

	//Schedule data saving in file with storeInterval
	ctx, cancel := context.WithCancel(context.Background())
	scheduleDataStoringInFile(ctx, &cfg, storage, &log)
	defer cancel()

	log.Info("Server started with Host setting: %s", cfg.Host)
	if err := http.ListenAndServe(cfg.Host, router); err != nil {
		panic(err)
	}
}

func scheduleDataStoringInFile(ctx context.Context, cfg *config.Config, storage *memstorage.MemStorage, log *logger.BaseLogger) *time.Ticker {
	var interval int64 = 1
	if cfg.StoreInterval > 1 {
		interval = cfg.StoreInterval
	}

	(*log).Info("[main::scheduleDataStoringInFile] init saving in file with interval: %d", cfg.StoreInterval)
	storeTicker := time.NewTicker(time.Duration(interval) * time.Second)
	go ticker.Run(storeTicker, ctx, func() { storage.SaveData() })

	return storeTicker
}

func createStorageManager(cfg *config.Config, log logger.BaseLogger) (storagemngr.StorageManager, error) {
	if cfg.DataBaseDSN != "" {
		return storagemngr.CreateDataBaseManager(cfg, log)
	} else if cfg.StoragePath != "" {
		return storagemngr.CreateFileManager(cfg.StoragePath, log), nil
	} else {
		return nil, nil
	}
}
