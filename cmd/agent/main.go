package main

import (
	"github.com/erupshis/metrics/internal/agent/agentimpl"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/ticker"
	"github.com/erupshis/metrics/internal/logger"
)

func main() {
	cfg := config.Parse()

	log := logger.CreateLogger(cfg.LogLevel)
	defer log.Sync()

	agent := agentimpl.Create(cfg, log)
	log.Info("Agent is started.")

	pollTicker := ticker.CreateWithSecondsInterval(agent.GetPollInterval())
	repeatTicker := ticker.CreateWithSecondsInterval(agent.GetReportInterval())

	defer pollTicker.Stop()
	defer repeatTicker.Stop()

	go ticker.Run(pollTicker, func() { agent.UpdateStats() })
	go ticker.Run(repeatTicker, func() { agent.PostJSONStats() })

	waitCh := make(chan struct{})
	<-waitCh
}
