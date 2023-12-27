// Package config implements agent's flags and environments parsing.
// Includes default param's for flags. Environments are more prioritized than flags.
package config

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/caarlos0/env"
	"github.com/erupshis/metrics/internal/configutils"
)

// Config stores agent's settings
//
// //go:generate easyjson -all config.go //DO NOT REGENERATE THIS FILE PollInterval & ReportInterval were modified in autogenerated file.
type Config struct {
	Host           string        `json:"address"`         // server's address
	PollInterval   time.Duration `json:"poll_interval"`   // stats collection interval
	ReportInterval time.Duration `json:"report_interval"` // sending stats on server interval
	RateLimit      int64         `json:"rate_limit"`      // number of simultaneous agent's connections to server
	Key            string        `json:"hash_key"`        // hash key for message check-su
	CertRSA        string        `json:"crypto_key"`      // CertRSA public certificate for connection.
	CACertRSA      string        `json:"ca_crypto_cert"`  // CertRSA public certificate for connection.
	RealIP         string        `json:"real_ip"`         // RealIP for client-server CIDR validation.
}

// ConfigDefault create default settings config. For debug use only.
var ConfigDefault = Config{
	Host:           "http://127.0.0.1",
	PollInterval:   2 * time.Second,
	ReportInterval: 10 * time.Second,
	RateLimit:      1,
	Key:            "123",
	CertRSA:        "rsa/cert.pem",
	CACertRSA:      "rsa/ca_cert.pem",
}

// Parse handling and reading settings from agent's launch flags and then environments,
// validates Host param and adds 'http://' prefix if missing.
func Parse() (Config, error) {
	var config = ConfigDefault
	if err := configutils.CheckConfigFile(&config); err != nil {
		return config, fmt.Errorf("parse config file: %w", err)
	}

	checkFlags(&config)

	if err := checkEnvironments(&config); err != nil {
		return config, fmt.Errorf("parse config: %w", err)
	}

	config.Host = configutils.AddHTTPPrefixIfNeed(config.Host)

	realIP, err := getRealIPAddr()
	if err != nil {
		return config, fmt.Errorf("real ip identification: %w", err)
	}

	config.RealIP = realIP
	return config, nil
}

const (
	flagAddress        = "a"
	flagReportInterval = "r"
	flagPollInterval   = "p"
	flagRateLimit      = "l"
	flagKey            = "k"
	flagCertRSA        = "crypto-key"    // flagCertRSA public connection key.
	flagCACertRSA      = "ca-crypto-key" // flagCACertRSA public connection ca cert.
)

func checkFlags(config *Config) {
	flag.StringVar(&config.Host, flagAddress, config.Host, "server host")
	flag.DurationVar(&config.ReportInterval, flagReportInterval, config.ReportInterval, "report interval val (sec)")
	flag.DurationVar(&config.PollInterval, flagPollInterval, config.PollInterval, "poll interval val (sec)")
	flag.Int64Var(&config.RateLimit, flagRateLimit, config.RateLimit, "rate limit")
	flag.StringVar(&config.Key, flagKey, config.Key, "auth key")
	flag.StringVar(&config.CertRSA, flagCertRSA, config.CertRSA, "public RSA key path")
	flag.StringVar(&config.CACertRSA, flagCACertRSA, config.CACertRSA, "public RSA CA cert path")
	flag.Parse()
}

// ENVIRONMENTS PARSING.
type envConfig struct {
	Host           string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
	RateLimit      string `env:"RATE_LIMIT"`
	Key            string `env:"KEY"`
	CertRSA        string `env:"CRYPTO_KEY"`    // CertRSA private key for connection.
	CACertRSA      string `env:"CA_CRYPTO_KEY"` // CertRSA private key for connection.
}

func checkEnvironments(config *Config) error {
	var envs = envConfig{}
	err := env.Parse(&envs)
	if err != nil {
		return fmt.Errorf("parse config environments: %w", err)
	}

	configutils.SetEnvToParamIfNeed(&config.Host, envs.Host)
	configutils.SetEnvToParamIfNeed(&config.RateLimit, envs.RateLimit)
	configutils.SetEnvToParamIfNeed(&config.ReportInterval, envs.ReportInterval)
	configutils.SetEnvToParamIfNeed(&config.PollInterval, envs.PollInterval)
	configutils.SetEnvToParamIfNeed(&config.Key, envs.Key)
	configutils.SetEnvToParamIfNeed(&config.CertRSA, envs.CertRSA)
	configutils.SetEnvToParamIfNeed(&config.CACertRSA, envs.CACertRSA)
	return nil
}

// getRealIPAddr Gets first non-local loop Network interface address.
func getRealIPAddr() (string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no ethernet adapter found")
}
