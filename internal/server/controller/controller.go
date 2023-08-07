package controller

import (
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/handlers"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/router"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	config   config.Config
	storage  *memstorage.MemStorage
	handlers *handlers.Handler
}

func Create() *Controller {
	storage := memstorage.Create()
	return &Controller{config.Parse(), storage, handlers.Create(storage)}
}

func (c *Controller) CreateRoutes() *chi.Mux {
	return router.Create(c.handlers)
}

func (c *Controller) GetConfig() *config.Config {
	return &c.config
}
