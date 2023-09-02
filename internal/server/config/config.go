package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
	"github.com/erupshis/metrics/internal/confighelper"
)

type Config struct {
	Host          string
	LogLevel      string
	Restore       bool
	StoreInterval int64
	StoragePath   string
	DataBaseDSN   string
}

func Parse() Config {
	var config = Config{}
	checkFlags(&config)
	checkEnvironments(&config)
	return config
}

// FLAGS PARSING.
const (
	flagAddress       = "a"
	flagLogLevel      = "l"
	flagRestore       = "r"
	flagStoragePath   = "f"
	flagStoreInterval = "i"
	flagDataBaseDSN   = "d"
)

func checkFlags(config *Config) {
	flag.StringVar(&config.Host, flagAddress, "localhost:8080", "server endpoint")
	flag.StringVar(&config.LogLevel, flagLogLevel, "Info", "log level")
	flag.BoolVar(&config.Restore, flagRestore, true, "restore values from file")

	//storagePathDef := "/tmp/metrics-db.json"
	flag.StringVar(&config.StoragePath, flagStoragePath, "", "file storage path")
	flag.Int64Var(&config.StoreInterval, flagStoreInterval, 5, "store interval val (sec)")

	//databaseDefDSN := "postgres://postgres:postgres@localhost:5432/metrics_bd?sslmode=disable"
	flag.StringVar(&config.DataBaseDSN, flagDataBaseDSN, "", "database DSN")
	flag.Parse()
}

// ENVIRONMENTS PARSING.
type envConfig struct {
	Host          string `env:"ADDRESS"`
	LogLevel      string `env:"LOG_LEVEL"`
	Restore       bool   `env:"RESTORE"`
	StoragePath   string `env:"FILE_STORAGE_PATH"`
	StoreInterval string `env:"STORE_INTERVAL"`
	DataBaseDSN   string `env:"DATABASE_DSN"`
}

func checkEnvironments(config *Config) {
	var envs = envConfig{}
	err := env.Parse(&envs)
	if err != nil {
		log.Fatal(err)
	}

	confighelper.SetEnvToParamIfNeed(&config.Host, envs.Host)
	confighelper.SetEnvToParamIfNeed(&config.LogLevel, envs.LogLevel)
	confighelper.SetEnvToParamIfNeed(&config.StoragePath, envs.StoragePath)
	confighelper.SetEnvToParamIfNeed(&config.StoreInterval, envs.StoreInterval)
	confighelper.SetEnvToParamIfNeed(&config.DataBaseDSN, envs.DataBaseDSN)

	config.Restore = envs.Restore || config.Restore
}
