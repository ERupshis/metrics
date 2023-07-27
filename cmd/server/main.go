package main

import (
	"github.com/ERupshis/macollection/internal/handlers"
	"net/http"
)

func run() error {
	customMux := http.NewServeMux()
	customMux.HandleFunc("/update/counter/", handlers.Counter)
	customMux.HandleFunc("/update/gauge/", handlers.Gauge)
	customMux.HandleFunc("/", handlers.Invalid)
	return http.ListenAndServe(`:8080`, customMux)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
