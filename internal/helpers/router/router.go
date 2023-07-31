package router

import (
	"github.com/ERupshis/metrics/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func Create() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handlers.ListHandler)
	r.Route("/update", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Get("/", handlers.MissingName)
			r.Post("/", handlers.MissingName)
			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", handlers.GetCounter)
				r.Post("/{value}", handlers.PostCounter)
			})

		})
		r.Route("/gauge", func(r chi.Router) {
			r.Get("/", handlers.MissingName)
			r.Post("/", handlers.MissingName)
			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", handlers.GetGauge)
				r.Post("/{value}", handlers.PostGauge)
			})
		})
	})

	r.NotFound(handlers.Invalid)

	return r
}
