package main

import (
	"github.com/ERupshis/metrics/internal/server/options"
	"github.com/ERupshis/metrics/internal/server/router"
	"net/http"
)

func main() {
	if err := http.ListenAndServe(options.Parse().Host, router.Create()); err != nil {
		panic(err)
	}
}
