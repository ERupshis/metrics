package main

import (
	"context"
	"time"

	"github.com/erupshis/metrics/internal/agent/agentimpl"
	"github.com/erupshis/metrics/internal/agent/client"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/ticker"
)

func main() {
	cfg := config.Parse()

	log := logger.CreateLogger(cfg.LogLevel)
	defer log.Sync()

	hash := hasher.CreateHasher(hasher.SHA256, log)
	defClient := client.CreateDefault(log, hash)

	agent := agentimpl.Create(cfg, log, defClient)
	log.Info("agent has started.")

	pollTicker := time.NewTicker(time.Duration(agent.GetPollInterval()) * time.Second)
	defer pollTicker.Stop()
	repeatTicker := time.NewTicker(time.Duration(agent.GetReportInterval()) * time.Second)
	defer repeatTicker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ticker.Run(pollTicker, ctx, func() { agent.UpdateStats() })
	go ticker.Run(repeatTicker, ctx, func() { agent.PostJSONStatsBatch(ctx) })

	waitCh := make(chan struct{})
	<-waitCh
}
