package main

import (
	. "github.com/ERupshis/metrics/internal/helpers/agentimpl"
	"time"
)

func main() {
	agent := CreateAgent()

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
