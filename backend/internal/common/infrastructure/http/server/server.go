package server

import (
	"context"
	"net/http"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
)

type Server struct {
	cfg *config.Config
	srv *http.Server
}

func New(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		cfg: cfg,
		srv: &http.Server{
			Addr:         cfg.Server.HTTP.Addr,
			Handler:      handler,
			ReadTimeout:  cfg.Server.Conns.ReadTimeout,
			WriteTimeout: cfg.Server.Conns.WriteTimeout,
			IdleTimeout:  cfg.Server.Conns.IdleTimeout,
		},
	}
}

func (s *Server) Run() error {
	if s.cfg.Server.HTTP.TLS.Enable {
		return s.srv.ListenAndServeTLS(
			s.cfg.Server.HTTP.TLS.ServerCertPath,
			s.cfg.Server.HTTP.TLS.ServerKeyPath,
		)
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
