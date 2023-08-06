package main

import (
	"github.com/ERupshis/metrics/internal/server/controller"
	"net/http"
)

func main() {
	serverController := controller.Create()
	if err := http.ListenAndServe(serverController.GetOptions().Host, serverController.CreateRoutes()); err != nil {
		panic(err)
	}
}
