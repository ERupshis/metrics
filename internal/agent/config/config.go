package config

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

type Config struct {
	Host           string
	ReportInterval int64
	PollInterval   int64
	LogLevel       string
}

func Default() Config {
	return Config{
		Host:           "http://localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
		LogLevel:       "Info",
	}
}

func Parse() Config {
	var config = Config{}
	checkFlags(&config)
	checkEnvironments(&config)
	config.Host = addHTTPPrefixIfNeed(config.Host)
	return config
}

// FLAGS PARSING.
const (
	flagAddress        = "a"
	flagReportInterval = "r"
	flagPollInterval   = "p"
	flagLogLevel       = "l"
)

func checkFlags(config *Config) {
	flag.StringVar(&config.Host, flagAddress, "http://localhost:8080", "server endpoint")
	flag.Int64Var(&config.ReportInterval, flagReportInterval, 10, "report interval val (sec)")
	flag.Int64Var(&config.PollInterval, flagPollInterval, 2, "poll interval val (sec)")
	flag.StringVar(&config.LogLevel, flagLogLevel, "Info", "log level")
	flag.Parse()
}

// ENVIRONMENTS PARSING.
type envConfig struct {
	Host           string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
	LogLevel       string `env:"LOG_LEVEL"`
}

func checkEnvironments(config *Config) {
	var envs = envConfig{}
	err := env.Parse(&envs)
	if err != nil {
		log.Fatal(err)
	}

	if envs.Host != "" {
		config.Host = envs.Host
	}

	if envs.ReportInterval != "" {
		if envVal, err := atoi64(envs.ReportInterval); err == nil {
			config.ReportInterval = envVal
		} else {
			panic(err)
		}
	}

	if envs.PollInterval != "" {
		if envVal, err := atoi64(envs.PollInterval); err == nil {
			config.PollInterval = envVal
		} else {
			panic(err)
		}
	}

	if envs.LogLevel != "" {
		config.LogLevel = envs.LogLevel
	}
}

// SUPPORT FUNCTIONS.
func atoi64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

//goland:noinspection HttpUrlsUsage
func addHTTPPrefixIfNeed(value string) string {
	if !strings.HasPrefix(value, "http://") {
		return "http://" + value
	}

	return value
}
