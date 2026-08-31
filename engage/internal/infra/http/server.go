package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/braunkc/hive-engage/config"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Router     *gin.Engine
	httpServer *http.Server
	log        *slog.Logger
}

func New(env config.Env, cfg config.HTTPConfig, log *slog.Logger) *Server {
	mode := gin.DebugMode
	if env != config.Debug {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	router := gin.New()

	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		err := fmt.Errorf("panic recovered: %v", recovered)
		log.Error("HTTP panic", slog.String("error", err.Error()), slog.Any("path", c.Request.URL.Path))

		if !c.Writer.Written() {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
	}))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	return &Server{
		Router:     router,
		httpServer: srv,
		log:        log,
	}
}

func (s *Server) Run() error {
	s.log.Info("starting http server...", slog.String("addr", s.httpServer.Addr))

	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server error: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("stopping http server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown error: %w", err)
	}

	return nil
}
