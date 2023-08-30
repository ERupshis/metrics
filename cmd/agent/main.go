package main

import (
	"context"

	"github.com/erupshis/metrics/internal/agent/agentimpl"
	"github.com/erupshis/metrics/internal/agent/client"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/ticker"
	"github.com/erupshis/metrics/internal/logger"
)

func main() {
	cfg := config.Parse()

	log := logger.CreateLogger(cfg.LogLevel)
	defer log.Sync()

	client := client.CreateDefault()

	agent := agentimpl.Create(cfg, log, client)
	log.Info("Agent is started.")

	pollTicker := ticker.CreateWithSecondsInterval(agent.GetPollInterval())
	repeatTicker := ticker.CreateWithSecondsInterval(agent.GetReportInterval())

	defer pollTicker.Stop()
	defer repeatTicker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ticker.Run(pollTicker, ctx, func() { agent.UpdateStats() })
	go ticker.Run(repeatTicker, ctx, func() { agent.PostJSONStats() })

	waitCh := make(chan struct{})
	<-waitCh
}
