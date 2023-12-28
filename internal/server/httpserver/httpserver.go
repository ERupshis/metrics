package httpserver

import (
	"context"
	"errors"
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
	info string
	host string
}

func NewServer(host string, router *chi.Mux, info string) *Server {
	srv := &http.Server{
		Addr:    host,
		Handler: router,
	}

	return &Server{
		Server: srv,
		info:   info,
	}
}

func (s *Server) Serve(lis net.Listener) error {
	if err := s.Server.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) GracefulStop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

func (s *Server) GetInfo() string {
	return s.info
}

func (s *Server) Host(host string) {
	s.host = host
}
func (s *Server) GetHost() string {
	return s.host
}
