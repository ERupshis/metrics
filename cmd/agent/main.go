package main

import (
	"github.com/erupshis/metrics/internal/agent/agentimpl"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/ticker"
)

func main() {
	agent := agentimpl.Create(config.Parse())

	pollTicker := ticker.CreateWithSecondsInterval(agent.GetPollInterval())
	repeatTicker := ticker.CreateWithSecondsInterval(agent.GetReportInterval())

	defer pollTicker.Stop()
	defer repeatTicker.Stop()

	go ticker.Run(pollTicker, func() { agent.UpdateStats() })
	go ticker.Run(repeatTicker, func() { agent.PostJSONStats() })

	waitCh := make(chan struct{})
	<-waitCh
}
