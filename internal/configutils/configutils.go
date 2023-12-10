// Package configutils provides functions for config handling.
package configutils

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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
