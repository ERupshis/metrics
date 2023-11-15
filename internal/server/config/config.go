// Package config provides configuration management for the metrics server.
// It defines a Config struct to hold various configuration parameters
// such as server address, log level, data restoration setting, storage path,
// store interval, database DSN, and authentication key.
// The package includes functions for parsing configuration from command line flags
// and environment variables, allowing flexibility in configuration.
// Additionally, utility functions from the internal configutils package are used
// to handle default values and set environment variables if needed.
package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
	"github.com/erupshis/metrics/internal/configutils"
)

// Config represents the configuration parameters for the metrics server.
type Config struct {
	Host          string // Host is the server endpoint (default: localhost:8080).
	LogLevel      string // LogLevel is the log level for the metrics server (default: Info).
	Restore       bool   // Restore enables or disables restoring values from a file (default: true).
	StoreInterval int64  // StoreInterval is the interval at which metrics are stored (default: 5 seconds).
	StoragePath   string // StoragePath is the file storage path for metrics data.
	DataBaseDSN   string // DataBaseDSN is the DSN for connecting to the metrics database.
	Key           string // Key is the authentication key for the metrics server.
}

// Parse reads and parses command line flags, updating the provided Config.
func Parse() Config {
	var config = Config{}
	checkFlags(&config)
	checkEnvironments(&config)
	return config
}

// FLAGS PARSING.

// Constants representing command line flags.
const (
	flagAddress       = "a" // flagAddress represents the server endpoint.
	flagLogLevel      = "l" // flagLogLevel represents the log level.
	flagRestore       = "r" // flagRestore represents the data restoration setting.
	flagStoragePath   = "f" // flagStoragePath represents the file storage path.
	flagStoreInterval = "i" // flagStoreInterval represents the store interval.
	flagDataBaseDSN   = "d" // flagDataBaseDSN represents the database DSN.
	flagKey           = "k" // flagKey represents the hash key.
)

// checkFlags initializes and parses command line flags, updating the provided Config.
func checkFlags(config *Config) {
	flag.StringVar(&config.Host, flagAddress, "localhost:8080", "server endpoint")
	flag.StringVar(&config.LogLevel, flagLogLevel, "Info", "log level")
	flag.BoolVar(&config.Restore, flagRestore, true, "restore values from file")

	// storagePathDef := "/tmp/metrics-db.json"
	flag.StringVar(&config.StoragePath, flagStoragePath, "", "file storage path")
	flag.Int64Var(&config.StoreInterval, flagStoreInterval, 5, "store interval val (sec)")

	databaseDefDSN := "postgres://postgres:postgres@localhost:5432/metrics_db?sslmode=disable"
	flag.StringVar(&config.DataBaseDSN, flagDataBaseDSN, databaseDefDSN, "database DSN")
	flag.StringVar(&config.Key, flagKey, "", "Auth key")
	flag.Parse()
}

// ENVIRONMENTS PARSING.

// envConfig represents the configuration parameters read from environment variables.
type envConfig struct {
	Host          string `env:"ADDRESS"`           // Host is the server endpoint.
	LogLevel      string `env:"LOG_LEVEL"`         // LogLevel is the log level.
	Restore       bool   `env:"RESTORE"`           // Restore is the data restoration setting.
	StoragePath   string `env:"FILE_STORAGE_PATH"` // StoragePath is the file storage path.
	StoreInterval string `env:"STORE_INTERVAL"`    // StoreInterval is the store interval.
	DataBaseDSN   string `env:"DATABASE_DSN"`      // DataBaseDSN is the database DSN.
	Key           string `env:"KEY"`               // Key is the hash key.
}

// checkEnvironments reads and parses environment variables, updating the provided Config.
func checkEnvironments(config *Config) {
	var envs = envConfig{}
	err := env.Parse(&envs)
	if err != nil {
		log.Fatal(err)
	}

	configutils.SetEnvToParamIfNeed(&config.Host, envs.Host)
	configutils.SetEnvToParamIfNeed(&config.LogLevel, envs.LogLevel)
	configutils.SetEnvToParamIfNeed(&config.StoragePath, envs.StoragePath)
	configutils.SetEnvToParamIfNeed(&config.StoreInterval, envs.StoreInterval)
	configutils.SetEnvToParamIfNeed(&config.DataBaseDSN, envs.DataBaseDSN)
	configutils.SetEnvToParamIfNeed(&config.Key, envs.Key)

	config.Restore = envs.Restore || config.Restore
}
