package main

import (
	"time"

	"github.com/erupshis/metrics/internal/agent/agentimpl"
	"github.com/erupshis/metrics/internal/agent/options"
)

func main() {
	agent := agentimpl.Create(options.Parse())

	var secondsFromStart int64
	secondsFromStart = 0
	for {
		time.Sleep(time.Second * 2)
		secondsFromStart += 2
		if secondsFromStart%agent.GetPollInterval() == 0 {
			agent.UpdateStats()
		}

		if secondsFromStart%agent.GetReportInterval() == 0 {
			agent.PostStats()
		}
	}
}
