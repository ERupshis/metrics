package server

import (
	"github.com/ERupshis/metrics/internal/helpers/router"
	"net/http"
)

func main() {
	if err := http.ListenAndServe(`:8080`, router.Create()); err != nil {
		panic(err)
	}
}
