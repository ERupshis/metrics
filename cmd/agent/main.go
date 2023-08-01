package main

import (
	"flag"
	"fmt"
	"github.com/ERupshis/metrics/internal/helpers/agentimpl"
	"os"
	"strings"
	"time"
)

var opts = agentimpl.Options{}

func parseFlags() {
	flag.StringVar(&opts.Host, "a", "localhost:8080", "server endpoint")
	flag.Int64Var(&opts.ReportInterval, "r", 10, "report interval val (sec)")
	flag.Int64Var(&opts.PollInterval, "p", 2, "poll interval val (sec)")
	flag.Parse()

	if len(opts.Host) == 0 {
		fmt.Println("empty arg a")
		os.Exit(1)
	} else if !strings.Contains(opts.Host, ":") {
		fmt.Println("missing port definition")
		os.Exit(1)
	}
}

func main() {
	parseFlags()

	agent := agentimpl.Create(opts)

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
