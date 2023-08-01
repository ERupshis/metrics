package main

import (
	"flag"
	"github.com/ERupshis/metrics/internal/helpers/router"
	"net/http"
)

type options struct {
	host string
}

var opt = options{}

func parseFlags() {
	flag.StringVar(&opt.host, "a", "localhost:8080", "server endpoint")
	flag.Parse()
}

func main() {
	parseFlags()

	if err := http.ListenAndServe(opt.host, router.Create()); err != nil {
		panic(err)
	}
}
