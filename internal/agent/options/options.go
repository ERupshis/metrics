package options

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
	"strconv"
	"strings"
)

type Options struct {
	Host           string
	ReportInterval int64
	PollInterval   int64
}

var flagAddress = "a"
var flagReportInterval = "r"
var flagPollInterval = "p"

type envConfig struct {
	Host           string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
}

func ParseOptions() Options {
	var opts = Options{}
	checkFlags(&opts)
	checkEnvironments(&opts)
	opts.Host = addHttpPrefixIfNeed(opts.Host)
	return opts
}

func checkFlags(opts *Options) {
	flag.StringVar(&opts.Host, flagAddress, "http://localhost:8080", "server endpoint")
	flag.Int64Var(&opts.ReportInterval, flagReportInterval, 10, "report interval val (sec)")
	flag.Int64Var(&opts.PollInterval, flagPollInterval, 2, "poll interval val (sec)")
	flag.Parse()
}

func checkEnvironments(opts *Options) {
	var envs = envConfig{}
	err := env.Parse(&envs)
	if err != nil {
		log.Fatal(err)
	}

	if envs.Host != "" {
		opts.Host = envs.Host
	}

	if envs.ReportInterval != "" {
		if envVal, err := atoi64(envs.ReportInterval); err == nil {
			opts.ReportInterval = envVal
		} else {
			panic(err)
		}
	}

	if envs.PollInterval != "" {
		if envVal, err := atoi64(envs.PollInterval); err == nil {
			opts.PollInterval = envVal
		} else {
			panic(err)
		}
	}
}

// /SUPPORT FUNCTIONS.
func atoi64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

//goland:noinspection HttpUrlsUsage
func addHttpPrefixIfNeed(value string) string {
	if !strings.Contains(value, "http://") {
		return "http://" + value
	}

	return value
}
