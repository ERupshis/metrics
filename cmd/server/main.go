package main

import (
	"flag"
	"fmt"
	"github.com/ERupshis/metrics/internal/helpers/router"
	"net/http"
	"os"
	"strings"
)

type options struct {
	host string
}

var opt = options{}

func parseFlags() {
	flag.StringVar(&opt.host, "a", "localhost:8080", "server endpoint")
	flag.Parse()

	if len(opt.host) == 0 {
		fmt.Println("empty arg a")
		os.Exit(1)
	} else if !strings.Contains(opt.host, ":") {
		fmt.Println("missing port definition")
		os.Exit(1)
	}
}

func main() {
	parseFlags()

	if err := http.ListenAndServe(opt.host, router.Create()); err != nil {
		panic(err)
	}
}
