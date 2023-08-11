package config

import (
	"flag"
	"os"
)

const (
	envAddress  = "ADDRESS"
	flagAddress = "a"
)

type Config struct {
	Host string
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
	flag.Parse()
}

// ENVIRONMENTS PARSING.
func checkEnvironments(config *Config) {
	if envHost := os.Getenv(envAddress); envHost != "" {
		config.Host = envHost
	}
}
