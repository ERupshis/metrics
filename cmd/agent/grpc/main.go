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
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/ticker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	log := logger.CreateLogger("info")
	defer log.Sync()

	IPparts := strings.Split(cfg.RealIP, "/")

	// TLS.
	creds, err := credentials.NewClientTLSFromFile(cfg.CertRSA, strings.Split(strings.TrimPrefix(cfg.Host, "http://"), ":")[0])
	if err != nil {
		log.Info("error create TLS cert: %v", err)
		return
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))
	opts = append(opts, grpc.WithUnaryInterceptor(logging.UnaryClient(log)))
	opts = append(opts, grpc.WithStreamInterceptor(logging.StreamClient(log)))

	grpcClient, err := client.CreateGRPC(strings.TrimPrefix(cfg.Host, "http://"), IPparts[0], opts...)
	if err != nil {
		log.Info("failed to create grpc client.")
		return
	}
	agent := agentimpl.Create(cfg, log, grpcClient)
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
