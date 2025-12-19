package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-backend-service/internal/config"
	"go-backend-service/internal/logger"

	"github.com/rs/zerolog"
)

// Server wraps the HTTP server and provides lifecycle methods
type Server struct {
	httpServer *http.Server
	log        zerolog.Logger
}

// New creates a new server instance
func New(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
		log: logger.Get(),
	}
}

// Start starts the server in a goroutine
func (s *Server) Start() error {
	go func() {
		s.log.Debug().
			Str("addr", s.httpServer.Addr).
			Dur("read_timeout", s.httpServer.ReadTimeout).
			Dur("write_timeout", s.httpServer.WriteTimeout).
			Dur("idle_timeout", s.httpServer.IdleTimeout).
			Msg("Starting HTTP server listener")

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Fatal().Err(err).Msg("HTTP server failed to start")
		}
	}()
	return nil
}

// Shutdown gracefully shuts down the server with a timeout
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info().
		Str("addr", s.httpServer.Addr).
		Msg("Initiating server shutdown")
	
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Error during server shutdown")
		return err
	}
	
	s.log.Info().Msg("Server shutdown completed successfully")
	return nil
}

// ShutdownWithTimeout gracefully shuts down the server with a default timeout
func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}
