package confighelper

import (
	"fmt"
	"strconv"
	"strings"
)

// SUPPORT FUNCTIONS.

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

func Atoi64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

//goland:noinspection HttpUrlsUsage
func AddHTTPPrefixIfNeed(value string) string {
	if !strings.HasPrefix(value, "http://") {
		return "http://" + value
	}

	return value
}
