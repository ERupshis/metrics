// Package configutils provides functions for config handling.
package configutils

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env"
)

// SetEnvToParamIfNeed assigns environment value to param depends on param's type definition.
// Accepts *int64, *string as params.
func SetEnvToParamIfNeed(param interface{}, val string) {
	if val == "" {
		return
	}

	switch param := param.(type) {
	case *int64:
		if envVal, err := Atoi64(val); err == nil {
			*param = envVal
		} else {
			panic(err)
		}
	case *time.Duration:
		if envVal, err := time.ParseDuration(val); err == nil {
			*param = envVal
		} else {
			panic(err)
		}
	case *string:
		*param = val
	default:
		panic(fmt.Errorf("wrong input param type"))
	}
}

// Atoi64 wrapper func to convert string into int64 if possible.
func Atoi64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

// AddHTTPPrefixIfNeed adds prefix 'http://' to string if missing.
//
//goland:noinspection HttpUrlsUsage
func AddHTTPPrefixIfNeed(value string) string {
	if !strings.HasPrefix(value, "http://") {
		return "http://" + value
	}

	return value
}

// CONFIG FILE PARSING.
const (
	flagConfigShort = "c"      // flagConfigShort path to config file short.
	flagConfig      = "config" // flagConfig path to config file.
)

type envFileConfig struct {
	Config string `env:"CONFIG"` // Config path to file config.
}

func CheckConfigFile(config any) error {
	var envs = envFileConfig{}
	err := env.Parse(&envs)
	if err != nil {
		return fmt.Errorf("parse config file path from env: %w", err)
	}

	configFilePath := ""
	SetEnvToParamIfNeed(&configFilePath, envs.Config)

	if configFilePath == "" {
		flag.StringVar(&configFilePath, flagConfig, "", "path to config file")
	}

	if configFilePath == "" {
		flag.StringVar(&configFilePath, flagConfigShort, "", "path to config file")
	}

	if configFilePath == "" {
		return nil
	}

	if err = ParseConfigFromFile(configFilePath, &config); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	return nil
}

func ParseConfigFromFile(filePath string, structToFill any) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if err = json.Unmarshal(data, structToFill); err != nil {
		return fmt.Errorf("parse data: %w", err)
	}

	return nil
}
