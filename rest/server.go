package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iden3/notification-service/config"
	"github.com/iden3/notification-service/log"
	"github.com/pkg/errors"
)

// Server app server
type Server struct {
	Routes     chi.Router
	httpServer *http.Server
	config     config.Server
}

// NewServer create new server
func NewServer(router chi.Router, c config.Server) *Server {
	return &Server{
		Routes: router,
		config: c,
	}
}

// Close closes connection to the server
func (s *Server) Close(ctx context.Context) error {
	return errors.WithStack(s.httpServer.Shutdown(ctx))
}

// Run server
func (s *Server) Run(port int) error {
	log.Infof("Server starting on port %d", port)
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           s.Routes,
		ReadHeaderTimeout: s.config.ReadHeaderTimeout,
		ReadTimeout:       s.config.ReadTimeout,
		IdleTimeout:       s.config.IdleTimeout,
		MaxHeaderBytes:    s.config.MaxHeaderBytes,
	}
	return errors.WithStack(s.httpServer.ListenAndServe())
}
