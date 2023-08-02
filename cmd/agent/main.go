package main

import (
	"flag"
	"github.com/ERupshis/metrics/internal/helpers/agentimpl"
	"github.com/caarlos0/env"
	"log"
	"strconv"
	"strings"
	"time"
)

type EnvConfig struct {
	Host           string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
}

func parseFlags() agentimpl.Options {
	var opts = agentimpl.Options{}
	flag.StringVar(&opts.Host, "a", "http://localhost:8080", "server endpoint")
	flag.Int64Var(&opts.ReportInterval, "r", 10, "report interval val (sec)")
	flag.Int64Var(&opts.PollInterval, "p", 2, "poll interval val (sec)")
	flag.Parse()

	if !strings.Contains(opts.Host, "http://") {
		opts.Host = "http://" + opts.Host
	}

	var envCfg EnvConfig
	err := env.Parse(&envCfg)
	if err != nil {
		log.Fatal(err)
	}

	if envCfg.Host != "" {
		opts.Host = envCfg.Host
	}

	if envCfg.ReportInterval != "" {
		if envVal, err := strconv.ParseInt(envCfg.ReportInterval, 10, 64); err == nil {
			opts.ReportInterval = envVal
		}
	}

	if envCfg.PollInterval != "" {
		if envVal, err := strconv.ParseInt(envCfg.PollInterval, 10, 64); err == nil {
			opts.PollInterval = envVal
		}
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
