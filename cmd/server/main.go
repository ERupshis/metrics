package main

import (
	"net/http"

	"github.com/erupshis/metrics/internal/server/controllers"
)

func main() {
	baseController := controllers.CreateBase()
	if err := http.ListenAndServe(baseController.GetConfig().Host, baseController.Route()); err != nil {
		panic(err)
	}
}
