package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
	"github.com/erupshis/metrics/internal/confighelper"
)

type Config struct {
	Host           string
	PollInterval   int64
	ReportInterval int64
	LogLevel       string
}

func Default() Config {
	return Config{
		Host:           "http://localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
		LogLevel:       "Info",
	}
}

func Parse() Config {
	var config = Config{}
	checkFlags(&config)
	checkEnvironments(&config)
	config.Host = confighelper.AddHTTPPrefixIfNeed(config.Host)
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

	confighelper.SetEnvToParamIfNeed(&config.Host, envs.Host)
	confighelper.SetEnvToParamIfNeed(&config.LogLevel, envs.LogLevel)
	confighelper.SetEnvToParamIfNeed(&config.ReportInterval, envs.ReportInterval)
	confighelper.SetEnvToParamIfNeed(&config.PollInterval, envs.PollInterval)
}
