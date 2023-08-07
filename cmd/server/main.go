package main

import (
	"net/http"

	"github.com/erupshis/metrics/internal/server/controller"
)

func main() {
	serverController := controller.Create()
	if err := http.ListenAndServe(serverController.GetOptions().Host, serverController.CreateRoutes()); err != nil {
		panic(err)
	}
}
