// Package configutils provides functions for config handling.
package configutils

import (
	"fmt"
	"strconv"
	"strings"
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
