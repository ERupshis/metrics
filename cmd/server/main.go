package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ipvalidatorGRPC "github.com/erupshis/metrics/internal/grpc/interceptors/ipvalidator"
	"github.com/erupshis/metrics/internal/grpc/interceptors/logging"
	"github.com/erupshis/metrics/internal/hasher"
	ipvalidatorHTTP "github.com/erupshis/metrics/internal/ipvalidator"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/rsa"
	"github.com/erupshis/metrics/internal/server"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/grpcserver"
	"github.com/erupshis/metrics/internal/server/grpcserver/controller"
	"github.com/erupshis/metrics/internal/server/httpserver"
	"github.com/erupshis/metrics/internal/server/httpserver/base"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
	"github.com/erupshis/metrics/internal/ticker"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

type serverInitializer struct {
	port     int64
	initFunc func(cfg *config.Config, log logger.BaseLogger, storage *memstorage.MemStorage) (server.BaseServer, error)
}

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
	storage := memstorage.Create(ctx, &cfg, storageManager, log)

	// Schedule data saving in file with storeInterval
	scheduleDataStoringInFile(ctx, &cfg, storage, log)

	// servers initializer
	var serversInitializer = map[string]serverInitializer{
		"http": serverInitializer{
			port:     cfg.PortHTTP,
			initFunc: initHTTPServer,
		},
		"grpc": serverInitializer{
			port:     cfg.PortGRPC,
			initFunc: initGRPCServer,
		},
	}

	// prepare servers if possible.
	var servers []server.BaseServer
	for serverType, initializer := range serversInitializer {
		if initializer.port == 0 {
			continue
		}

		srv, err := initializer.initFunc(&cfg, log, storage)
		if err != nil {
			log.Info("failed to init %s server: %v", serverType, err)
		} else {
			servers = append(servers, srv)
		}
	}

	// launch servers.
	var wg sync.WaitGroup
	idleConnsClosed := make(chan struct{})
	for _, srv := range servers {
		if err := launchServer(&wg, idleConnsClosed, srv, log); err != nil {
			log.Info("failed to start %s server: %v", srv.GetInfo(), err)
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

func createHTTPTrustedSubnetValidator(cfg *config.Config, log logger.BaseLogger) *ipvalidatorHTTP.ValidatorIP {
	_, subnet, err := net.ParseCIDR(cfg.TrustedSubnet)
	if err != nil {
		log.Info("[main:createHTTPTrustedSubnetValidator] failed to parse CIDR: %v", err)
	}

	return ipvalidatorHTTP.Create(subnet)
}

func createGRPCTrustedSubnetValidator(cfg *config.Config, log logger.BaseLogger) *ipvalidatorGRPC.ValidatorIP {
	_, subnet, err := net.ParseCIDR(cfg.TrustedSubnet)
	if err != nil {
		log.Info("[main:createGRPCTrustedSubnetValidator] failed to parse CIDR: %v", err)
	}

	return ipvalidatorGRPC.Create(subnet, "")
}

func initShutDown(ctx context.Context, idleConnsClosed chan struct{}, servers []server.BaseServer, logger logger.BaseLogger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigCh
		logger.Info("[main:initShutDown] application is stopping gracefully")
		for _, srv := range servers {
			if err := srv.GracefulStop(ctx); err != nil {
				logger.Info("[main:initShutDown] %s server graceful stop error: %v", srv.GetInfo(), err)
			}
		}
		close(idleConnsClosed)
	}()
}

func initHTTPServer(cfg *config.Config, log logger.BaseLogger, storage *memstorage.MemStorage) (server.BaseServer, error) {
	// hash sum evaluation
	hash := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// rsa encrypting
	rsaDecoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		return nil, fmt.Errorf("[main:initHTTPServer] failed to create RSA decoder: %v", err)
	}

	// trusted subnet validation.
	validatorIP := createHTTPTrustedSubnetValidator(cfg, log)

	baseController := base.Create(cfg, log, storage, hash, rsaDecoder, validatorIP)

	router := chi.NewRouter()
	router.Mount("/", baseController.Route())

	// server launch.
	srv := httpserver.NewServer(cfg.Host, router, "http")
	srv.Host(fmt.Sprintf("%s:%d", cfg.Host, cfg.PortHTTP))
	return srv, nil
}

func initGRPCServer(cfg *config.Config, log logger.BaseLogger, storage *memstorage.MemStorage) (server.BaseServer, error) {
	grpcController := controller.New(storage)
	// trusted subnet validation.
	validatorIP := createGRPCTrustedSubnetValidator(cfg, log)

	// TLS.
	cert, err := tls.LoadX509KeyPair(cfg.CertRSA, cfg.KeyRSA)
	if err != nil {
		return nil, fmt.Errorf("error create TLS cert: %w", err)
	}

	// gRPC server options.
	var opts []grpc.ServerOption
	opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
	opts = append(opts, grpc.ChainUnaryInterceptor(
		logging.UnaryServer(log),
		validatorIP.UnaryServer(log),
	))
	opts = append(opts, grpc.ChainStreamInterceptor(
		logging.StreamServer(log),
		validatorIP.StreamServer(log),
	))

	srv := grpcserver.NewServer(grpcController, "grpc", opts...)
	srv.Host(fmt.Sprintf(":%d", cfg.PortGRPC))
	return srv, nil
}

func launchServer(wg *sync.WaitGroup, idleConnsClosed <-chan struct{}, srv server.BaseServer, log logger.BaseLogger) error {
	log.Info("%s server is launching with Host setting: %s", srv.GetInfo(), srv.GetHost())

	listener, err := net.Listen("tcp", srv.GetHost())
	if err != nil {
		return fmt.Errorf("failed to listen for %s server: %w", srv.GetInfo(), err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err = srv.Serve(listener); err != nil {
			log.Info("%s server refused to start or stop with error: %v", srv.GetInfo(), err)
			return
		}

		<-idleConnsClosed
		log.Info("%s server shutdown gracefully", srv.GetInfo())
	}()

	return nil
}
