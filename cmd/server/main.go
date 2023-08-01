package main

import (
	"flag"
	"github.com/ERupshis/metrics/internal/helpers/router"
	"net/http"
)

type options struct {
	host string
}

func parseFlags() options {
	var opts = options{}
	flag.StringVar(&opts.host, "a", "localhost:8080", "server endpoint")
	flag.Parse()

	return opts
}

func main() {
	if err := http.ListenAndServe(parseFlags().host, router.Create()); err != nil {
		panic(err)
	}
}
