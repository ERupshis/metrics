package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/erupshis/metrics/internal/agent/agentimpl"
	"github.com/erupshis/metrics/internal/agent/client"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/workers"
	"github.com/erupshis/metrics/internal/grpc/interceptors/logging"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/rsa"
	"github.com/erupshis/metrics/internal/ticker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

type agentClientInitializer struct {
	initFunc func(cfg *config.Config, log logger.BaseLogger) (client.BaseClient, error)
}

func main() {
	// example of run: go run -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(cmd.exe /c "echo %DATE%")' -X 'main.buildCommit=$(git rev-parse HEAD)'" main.go
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("error parse config: %v", err)
		return
	}

	log := logger.CreateLogger("info")
	defer log.Sync()

	var clientsInitializer = map[string]agentClientInitializer{
		"http": agentClientInitializer{
			initFunc: initHTTPClient,
		},
		"grpc": agentClientInitializer{
			initFunc: initGRPCClient,
		},
	}

	clientInitializer, ok := clientsInitializer[cfg.ClientType]
	if !ok {
		log.Info("unknown client type, cannot proceed")
		return
	}

	agentClient, err := clientInitializer.initFunc(&cfg, log)
	if err != nil {
		log.Info("failed to create client: %v", err)
		return
	}

	agent := agentimpl.Create(cfg, log, agentClient)
	log.Info("agent has started.")

	pollTicker := time.NewTicker(agent.GetPollInterval())
	defer pollTicker.Stop()
	repeatTicker := time.NewTicker(agent.GetReportInterval())
	defer repeatTicker.Stop()

	ctx, cancel := context.WithCancel(context.Background())

	workersPool, err := workers.CreateWorkersPool(cfg.RateLimit, log)
	if err != nil {
		log.Info("failed to create workers.")
		return
	}
	defer workersPool.CloseJobsChan()
	defer workersPool.CloseResultsChan()

	go ticker.Run(pollTicker, ctx, func() { agent.UpdateStats() })
	go ticker.Run(pollTicker, ctx, func() { agent.UpdateExtraStats() })
	go ticker.Run(repeatTicker, ctx, func() { go workersPool.AddJob(func() error { return agent.PostStatsBatch(ctx) }) })

	go func() {
		for res := range workersPool.GetResultChan() {
			if res != nil {
				log.Info("[WorkersPool] failed work: %v", res)
			}
		}
	}()

	idleConnsClosed := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigCh
		cancel()
		close(idleConnsClosed)
	}()

	<-idleConnsClosed
}

func initHTTPClient(cfg *config.Config, log logger.BaseLogger) (client.BaseClient, error) {
	// hash sum evaluation
	hash := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// rsa encrypting
	rsaEncoder, err := rsa.CreateEncoder(cfg.CertRSA)
	if err != nil {
		log.Info("[main] failed to create RSA encoder: %v", err)
	}

	IPparts := strings.Split(cfg.RealIP, "/")
	return client.CreateDefault(log, hash, rsaEncoder, IPparts[0], cfg.Host), nil
}

func initGRPCClient(cfg *config.Config, log logger.BaseLogger) (client.BaseClient, error) {
	IPparts := strings.Split(cfg.RealIP, "/")

	// TLS.
	serverAddressWOPrefix := strings.TrimPrefix(cfg.Host, "http://")
	creds, err := credentials.NewClientTLSFromFile(cfg.CertRSA, strings.Split(serverAddressWOPrefix, ":")[0])
	if err != nil {
		return nil, fmt.Errorf("error create TLS cert: %w", err)
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))
	opts = append(opts, grpc.WithChainUnaryInterceptor(
		logging.UnaryClient(log),
	))
	opts = append(opts, grpc.WithChainStreamInterceptor(
		logging.StreamClient(log),
	))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))

	return client.CreateGRPC(serverAddressWOPrefix, IPparts[0], opts...)
}
