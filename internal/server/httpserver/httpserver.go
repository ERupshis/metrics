package httpserver

import (
	"context"
	"net"
	"net/http"

	"github.com/erupshis/metrics/internal/server"
	"github.com/go-chi/chi/v5"
)

var (
	_ server.BaseServer = (*Server)(nil)
)

type Server struct {
	*http.Server
}

func NewServer(host string, router *chi.Mux) *Server {
	srv := &http.Server{
		Addr:    host,
		Handler: router,
	}

	return &Server{
		Server: srv,
	}
}

func (s *Server) Serve(lis net.Listener) error {
	return s.Server.Serve(lis)
}

func (s *Server) GracefulStop(ctx context.Context) error {
	return s.Shutdown(ctx)
}
