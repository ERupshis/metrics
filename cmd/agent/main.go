package main

import (
	"flag"
	"github.com/ERupshis/metrics/internal/helpers/agentimpl"
	"strings"
	"time"
)

func parseFlags() agentimpl.Options {
	var opts = agentimpl.Options{}
	flag.StringVar(&opts.Host, "a", "http://localhost:8080", "server endpoint")
	flag.Int64Var(&opts.ReportInterval, "r", 10, "report interval val (sec)")
	flag.Int64Var(&opts.PollInterval, "p", 2, "poll interval val (sec)")
	flag.Parse()

	if !strings.Contains(opts.Host, "opts.Host") {
		opts.Host = "http://" + opts.Host
	}
	return opts
}

func main() {

	agent := agentimpl.Create(parseFlags())

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
