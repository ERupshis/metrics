package main

import (
	"net/http"

	"github.com/erupshis/metrics/internal/server/controllers"
	"github.com/go-chi/chi/v5"
)

func main() {
	baseController := controllers.CreateBase()

	router := chi.NewRouter()
	router.Mount("/", baseController.Route())

	if err := http.ListenAndServe(baseController.GetConfig().Host, router); err != nil {
		panic(err)
	}
}
