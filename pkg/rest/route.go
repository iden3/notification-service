package rest

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	"github.com/iden3/signed-notification/pkg/rest/handlers"
)

// Handlers server handlers
type Handlers struct {
	proxyHandler *handlers.ProxyHandler
	keyHandler   *handlers.KeyHandler
}

// NewHandlers create handlers.
func NewHandlers(p *handlers.ProxyHandler, k *handlers.KeyHandler) *Handlers {
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

	r.Post("/_matrix/push/v1/notify", s.proxyHandler.ProxyNotification)
	r.Get("/public_key", s.keyHandler.GetPublicKey)

	return r
}
