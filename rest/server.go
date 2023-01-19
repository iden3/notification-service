package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"

	"github.com/iden3/notification-service/log"
)

// Server app server
type Server struct {
	Routes     chi.Router
	httpServer *http.Server
}

// NewServer create new server
func NewServer(router chi.Router) *Server {
	return &Server{
		Routes: router,
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
		ReadHeaderTimeout: time.Second,
	}
	return errors.WithStack(s.httpServer.ListenAndServe())
}
