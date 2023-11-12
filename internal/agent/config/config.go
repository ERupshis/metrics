// Package config implements agent's flags and environments parsing.
// Includes default param's for flags. Environments are more prioritized than flags.
package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
	"github.com/erupshis/metrics/internal/confighelper"
)

// Config stores agent's settings
type Config struct {
	Host           string // server's address
	PollInterval   int64  // stats collection interval
	ReportInterval int64  // sending stats on server interval
	RateLimit      int64  // number of simultaneous agent's connections to server
	Key            string // hash key for message check-su
}

// Default create default settings config. For debug use only.
func Default() Config {
	return Config{
		Host:           "http://localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
		RateLimit:      1,
		Key:            "123",
	}
}

// Parse handling and reading settings from agent's launch flags and then environments,
// validates Host param and adds 'http://' prefix if missing.
func Parse() Config {
	var config = Config{}
	checkFlags(&config)
	checkEnvironments(&config)
	config.Host = confighelper.AddHTTPPrefixIfNeed(config.Host)
	return config
}

const (
	flagAddress        = "a"
	flagReportInterval = "r"
	flagPollInterval   = "p"
	flagRateLimit      = "l"
	flagKey            = "k"
)

func checkFlags(config *Config) {
	flag.StringVar(&config.Host, flagAddress, "http://localhost:8080", "server endpoint")
	flag.Int64Var(&config.ReportInterval, flagReportInterval, 10, "report interval val (sec)")
	flag.Int64Var(&config.PollInterval, flagPollInterval, 2, "poll interval val (sec)")
	flag.Int64Var(&config.RateLimit, flagRateLimit, 1, "rate limit")
	flag.StringVar(&config.Key, flagKey, "", "auth key")
	flag.Parse()
}

// ENVIRONMENTS PARSING.
type envConfig struct {
	Host           string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
	RateLimit      string `env:"RATE_LIMIT"`
	Key            string `env:"KEY"`
}

func checkEnvironments(config *Config) {
	var envs = envConfig{}
	err := env.Parse(&envs)
	if err != nil {
		log.Fatal(err)
	}

	confighelper.SetEnvToParamIfNeed(&config.Host, envs.Host)
	confighelper.SetEnvToParamIfNeed(&config.RateLimit, envs.RateLimit)
	confighelper.SetEnvToParamIfNeed(&config.ReportInterval, envs.ReportInterval)
	confighelper.SetEnvToParamIfNeed(&config.PollInterval, envs.PollInterval)
	confighelper.SetEnvToParamIfNeed(&config.Key, envs.Key)
}
