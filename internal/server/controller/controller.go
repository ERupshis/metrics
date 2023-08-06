package controller

import (
	"github.com/ERupshis/metrics/internal/server/handlers"
	"github.com/ERupshis/metrics/internal/server/memstorage"
	"github.com/ERupshis/metrics/internal/server/options"
	"github.com/ERupshis/metrics/internal/server/router"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	options  options.Options
	storage  *memstorage.MemStorage
	handlers *handlers.Handler
}

func Create() *Controller {
	storage := memstorage.Create()
	return &Controller{options.Parse(), storage, handlers.Create(storage)}
}

func (c *Controller) CreateRoutes() *chi.Mux {
	return router.Create(c.handlers)
}

func (c *Controller) GetOptions() *options.Options {
	return &c.options
}
