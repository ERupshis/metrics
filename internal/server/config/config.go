package config

import (
	"flag"
	"os"
)

const (
	envAddress  = "ADDRESS"
	envLogLevel = "LOG_LEVEL"

	flagAddress  = "a"
	flagLogLevel = "l"
)

type Config struct {
	Host     string
	LogLevel string
}

func Parse() Config {
	var config = Config{}
	checkFlags(&config)
	checkEnvironments(&config)
	return config
}

// FLAGS PARSING.
func checkFlags(config *Config) {
	flag.StringVar(&config.Host, flagAddress, "localhost:8080", "server endpoint")
	flag.StringVar(&config.LogLevel, flagLogLevel, "Info", "log level")
	flag.Parse()
}

// ENVIRONMENTS PARSING.
func checkEnvironments(config *Config) {
	if envHost := os.Getenv(envAddress); envHost != "" {
		config.Host = envHost
	}

	if envLogLvl := os.Getenv(envLogLevel); envLogLvl != "" {
		config.LogLevel = envLogLvl
	}
}
