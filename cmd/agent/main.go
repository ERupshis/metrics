package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erupshis/metrics/internal/agent/agentimpl"
	"github.com/erupshis/metrics/internal/agent/client"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/workers"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	rsa "github.com/erupshis/metrics/internal/rsa"
	"github.com/erupshis/metrics/internal/ticker"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	// example of run: go run -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(cmd.exe /c "echo %DATE%")' -X 'main.buildCommit=$(git rev-parse HEAD)'" main.go
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	cfg := config.Parse()

	log := logger.CreateLogger("info")
	defer log.Sync()

	// hash sum evaluation
	hash := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// rsa encrypting
	rsaEncoder, err := rsa.CreateEncoder(cfg.CertRSA)
	if err != nil {
		log.Info("[main] failed to create RSA encoder: %v", err)
	}

	defClient := client.CreateDefault(log, hash, rsaEncoder)

	agent := agentimpl.Create(cfg, log, defClient)
	log.Info("agent has started.")

	pollTicker := time.NewTicker(time.Duration(agent.GetPollInterval()) * time.Second)
	defer pollTicker.Stop()
	repeatTicker := time.NewTicker(time.Duration(agent.GetReportInterval()) * time.Second)
	defer repeatTicker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workersPool, err := workers.CreateWorkersPool(cfg.RateLimit, log)
	if err != nil {
		log.Info("failed to create workers.")
		return
	}
	defer workersPool.CloseJobsChan()
	defer workersPool.CloseResultsChan()

	go ticker.Run(pollTicker, ctx, func() { agent.UpdateStats() })
	go ticker.Run(pollTicker, ctx, func() { agent.UpdateExtraStats() })
	go ticker.Run(repeatTicker, ctx, func() { go workersPool.AddJob(func() error { return agent.PostJSONStatsBatch(ctx) }) })

	go func() {
		for res := range workersPool.GetResultChan() {
			if res != nil {
				log.Info("[WorkersPool] failed work: %v", res)
			}
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
