package main

import (
	"context"
	"time"

	"github.com/erupshis/metrics/internal/agent/agentimpl"
	"github.com/erupshis/metrics/internal/agent/client"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/workers"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/ticker"
)

func main() {
	cfg := config.Parse()

	log := logger.CreateLogger("info")
	defer log.Sync()

	hash := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)
	defClient := client.CreateDefault(log, hash)

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

	for res := range workersPool.GetResultChan() {
		if res != nil {
			log.Info("[WorkersPool] failed work: %v", res)
		}
	}
}
