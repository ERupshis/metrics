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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storageManager := createStorageManager(ctx, &cfg, log)
	if storageManager != nil {
		defer storageManager.Close()
	}
	storage := memstorage.Create(storageManager)
	baseController := controllers.CreateBase(ctx, cfg, log, storage)

	router := chi.NewRouter()
	router.Mount("/", baseController.Route())

	//Schedule data saving in file with storeInterval
	scheduleDataStoringInFile(ctx, &cfg, storage, log)

	log.Info("Server started with Host setting: %s", cfg.Host)
	if err := http.ListenAndServe(cfg.Host, router); err != nil {
		log.Info("server refused to start with error: %v", err)
		panic(err)
	}
}

func scheduleDataStoringInFile(ctx context.Context, cfg *config.Config, storage *memstorage.MemStorage, log logger.BaseLogger) *time.Ticker {
	var interval int64 = 1
	if cfg.StoreInterval > 1 {
		interval = cfg.StoreInterval
	}

	log.Info("[main::scheduleDataStoringInFile] init saving in file with interval: %d", cfg.StoreInterval)
	storeTicker := time.NewTicker(time.Duration(interval) * time.Second)
	go ticker.Run(storeTicker, ctx, func() {
		err := storage.SaveData(ctx)
		if err != nil {
			log.Info("[main::scheduleDataStoringInFile] failed to save data, error: %v", err)
		}
	})

	return storeTicker
}

func createStorageManager(ctx context.Context, cfg *config.Config, log logger.BaseLogger) storagemngr.StorageManager {
	if cfg.DataBaseDSN != "" {
		manager, err := storagemngr.CreateDataBaseManager(ctx, cfg, log)
		if err != nil {
			log.Info("[main] failed to create connection to database: %s with error: %v", cfg.DataBaseDSN, err)
		}
		return manager
	} else if cfg.StoragePath != "" {
		return storagemngr.CreateFileManager(cfg.StoragePath, log)
	} else {
		return nil
	}
}
