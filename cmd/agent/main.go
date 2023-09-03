package main

import (
	"context"
	"time"

	"github.com/erupshis/metrics/internal/agent/agentimpl"
	"github.com/erupshis/metrics/internal/agent/client"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/ticker"
)

func main() {
	cfg := config.Parse()

	log := logger.CreateLogger(cfg.LogLevel)
	defer log.Sync()

	client := client.CreateDefault()

	agent := agentimpl.Create(cfg, log, client)
	log.Info("Agent is started.")

	pollTicker := time.NewTicker(time.Duration(agent.GetPollInterval()) * time.Second)
	defer pollTicker.Stop()
	repeatTicker := time.NewTicker(time.Duration(agent.GetReportInterval()) * time.Second)
	defer repeatTicker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ticker.Run(pollTicker, ctx, func() { agent.UpdateStats() })
	go ticker.Run(repeatTicker, ctx, func() { _ = agent.PostJSONStatsBatch() })

	waitCh := make(chan struct{})
	<-waitCh
}
