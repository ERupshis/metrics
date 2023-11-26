package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/controllers"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
	"github.com/erupshis/metrics/internal/ticker"
	"github.com/go-chi/chi/v5"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	// example of run: go run -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(cmd.exe /c "echo %DATE%")' -X 'main.buildCommit=$(git rev-parse HEAD)'" main.go
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	/*fcpu, err := os.Create(`profiles/cpu_optimized_logger_removed_from_SaveMetrics.pprof`)
	if err != nil {
		panic(err)
	}
	defer fcpu.Close()
	if err := pprof.StartCPUProfile(fcpu); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()*/

	cfg := config.Parse()

	log := logger.CreateLogger(cfg.LogLevel)
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storageManager := createStorageManager(ctx, &cfg, log)
	if storageManager != nil {
		defer func() {
			if err := storageManager.Close(); err != nil {
				log.Info("failed to close storage: %v", err)
			}
		}()
	}
	storage := memstorage.Create(storageManager)
	hash := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)
	baseController := controllers.CreateBase(ctx, cfg, log, storage, hash)

	router := chi.NewRouter()
	router.Mount("/", baseController.Route())

	// Schedule data saving in file with storeInterval
	scheduleDataStoringInFile(ctx, &cfg, storage, log)

	// heap profiling.
	// router.Mount("/debug", middleware.Profiler())

	// server launch.
	go func() {
		log.Info("server is launching with Host setting: %s", cfg.Host)
		if err := http.ListenAndServe(cfg.Host, router); err != nil {
			log.Info("server refused to start with error: %v", err)
		}
	}()

	// time.Sleep(300 * time.Second)
	// memProfile()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
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
			log.Info("[main:createStorageManager] failed to create connection to database: %s with error: %v", cfg.DataBaseDSN, err)
		}
		return manager
	} else if cfg.StoragePath != "" {
		return storagemngr.CreateFileManager(cfg.StoragePath, log)
	} else {
		return nil
	}
}

/*func memProfile() {
	// создаём файл журнала профилирования памяти
	fmem, err := os.Create(`profiles/result.pprof`)
	if err != nil {
		panic(err)
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := pprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}
}*/
