package main

import (
	"context"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/ipvalidator"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/rsa"
	"github.com/erupshis/metrics/internal/server"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/httpserver"
	"github.com/erupshis/metrics/internal/server/httpserver/base"
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

	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("error parse config: %v", err)
		return
	}

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

	// Schedule data saving in file with storeInterval
	scheduleDataStoringInFile(ctx, &cfg, storage, log)

	idleConnsClosed := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	var servers []server.BaseServer
	if cfg.PortHTTP != 0 {
		httpServer, err := launchHTTPServer(ctx, &wg, idleConnsClosed, &cfg, log, storage)
		if err != nil {
			log.Info("failed to start http server: %v", err)
		} else {
			servers = append(servers, httpServer)
		}
	}

	if cfg.PortGRPC != 0 {
		grpcServer, err := launchHTTPServer(ctx, &wg, idleConnsClosed, &cfg, log, storage)
		if err != nil {
			log.Info("failed to start grpc server: %v", err)
		} else {
			servers = append(servers, grpcServer)
		}
	}

	initShutDown(ctx, idleConnsClosed, servers, log)
	wg.Wait()
}

func scheduleDataStoringInFile(ctx context.Context, cfg *config.Config, storage *memstorage.MemStorage, log logger.BaseLogger) *time.Ticker {
	interval := time.Second
	if cfg.StoreInterval > 1 {
		interval = cfg.StoreInterval
	}

	log.Info("[main::scheduleDataStoringInFile] init saving in file with interval: %s", cfg.StoreInterval.String())
	storeTicker := time.NewTicker(interval)
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

func createTrustedSubnetValidator(cfg *config.Config, log logger.BaseLogger) *ipvalidator.ValidatorIP {
	_, subnet, err := net.ParseCIDR(cfg.TrustedSubnet)
	if err != nil {
		log.Info("[main:createTrustedSubnetValidator] failed to parse CIDR: %v", err)
	}

	return ipvalidator.Create(subnet)
}

func initShutDown(ctx context.Context, idleConnsClosed chan struct{}, servers []server.BaseServer, logger logger.BaseLogger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigCh
		for _, srv := range servers {
			logger.Info("http server is going to gracefully shutdown")
			if err := srv.GracefulStop(ctx); err != nil {
				logger.Info("graceful stop error: %v", err)
			}
		}
		close(idleConnsClosed)
	}()
}

func launchHTTPServer(ctx context.Context,
	wg *sync.WaitGroup,
	idleConnsClosed <-chan struct{},
	cfg *config.Config,
	log logger.BaseLogger,
	storage *memstorage.MemStorage) (server.BaseServer, error) {
	// hash sum evaluation
	hash := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// rsa encrypting
	rsaDecoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("[main] failed to create RSA decoder: %v", err)
	}

	// trusted subnet validation.
	validatorIP := createTrustedSubnetValidator(cfg, log)

	baseController := base.Create(ctx, cfg, log, storage, hash, rsaDecoder, validatorIP)

	router := chi.NewRouter()
	router.Mount("/", baseController.Route())

	// server launch.
	mainServer := httpserver.NewServer(cfg.Host, router)

	log.Info("server is launching with Host setting: %s", fmt.Sprintf("%s:%d", cfg.Host, cfg.PortHTTP))

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.PortHTTP))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		defer wg.Done()

		if err = mainServer.Serve(listener); err != nil {
			log.Info("http server refused to start or stop with error: %v", err)
		}

		<-idleConnsClosed
		log.Info("http server shutdown gracefully")
	}()

	return mainServer, nil
}
