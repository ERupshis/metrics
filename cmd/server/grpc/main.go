package main

import (
	"context"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erupshis/metrics/internal/grpc/interceptors/logging"
	"github.com/erupshis/metrics/internal/ipvalidator"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/grpcserver"
	"github.com/erupshis/metrics/internal/server/grpcserver/controller"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
	"github.com/erupshis/metrics/internal/ticker"
	"google.golang.org/grpc"
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

	// trusted subnet validation.
	_ = createTrustedSubnetValidator(&cfg, log)

	grpcController := controller.New(storage)

	// Schedule data saving in file with storeInterval
	scheduleDataStoringInFile(ctx, &cfg, storage, log)

	// gRPC server options.
	var opts []grpc.ServerOption
	opts = append(opts, grpc.ChainUnaryInterceptor(logging.UnaryServer(log)))
	opts = append(opts, grpc.ChainStreamInterceptor(logging.StreamServer(log)))

	grpcServer := grpcserver.NewServer(grpcController, log, opts...)
	idleConnsClosed := initShutDown(grpcServer, log)

	_, port, err := net.SplitHostPort(cfg.Host)
	if err != nil {
		log.Info("failed to parse port for gRPC")
		return
	}

	log.Info("starting gRPC listener on port %s", port)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Info("failed to listen: %v", err)
		return
	}

	if err = grpcServer.Serve(listener); err != nil {
		log.Info("gRPC server refused to start or stop with error: %v", err)
	}

	<-idleConnsClosed
	log.Info("gRPC server shutdown gracefully")
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
		if err != nil {
			log.Info("[main:createTrustedSubnetValidator] failed to parse CIDR: %v", err)
		}
	}

	return ipvalidator.Create(subnet)
}

func initShutDown(srv *grpcserver.Server, logger logger.BaseLogger) <-chan struct{} {
	idleConnsClosed := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigCh
		logger.Info("gRPC server is going to gracefully shutdown")
		srv.GracefulStop()
		close(idleConnsClosed)
	}()

	return idleConnsClosed
}
