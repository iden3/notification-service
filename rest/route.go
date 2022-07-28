package rest

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	"github.com/iden3/notification-service/rest/handlers"
)

// Handlers server handlers
type Handlers struct {
	proxyHandler *handlers.PushNotificationHandler
	keyHandler   *handlers.KeyHandler
}

// NewHandlers create handlers.
func NewHandlers(p *handlers.PushNotificationHandler, k *handlers.KeyHandler) *Handlers {
	return &Handlers{
		proxyHandler: p,
		keyHandler:   k,
	}
}

// Routes chi http routs configuration
func (s *Handlers) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, struct {
			Status string `json:"status"`
		}{Status: "up and running"})
	})

	r.Post("/api/v1/", s.proxyHandler.Send)
	r.Get("/api/v1/{id}", s.proxyHandler.Get)

	r.Get("/api/v1/public", s.keyHandler.GetPublicKey)

	return r
}
