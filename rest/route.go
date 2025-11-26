package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/iden3/notification-service/config"
	"github.com/iden3/notification-service/rest/handlers"
)

// Handlers server handlers
type Handlers struct {
	proxyHandler *handlers.PushNotificationHandler
	keyHandler   *handlers.KeyHandler

	authmiddleware func(http.Handler) http.Handler
	corsCfg        config.CORS
}

// NewHandlers create handlers.
func NewHandlers(
	p *handlers.PushNotificationHandler,
	k *handlers.KeyHandler,
	a func(http.Handler) http.Handler,
	corsCfg config.CORS) *Handlers {
	return &Handlers{
		proxyHandler:   p,
		keyHandler:     k,
		authmiddleware: a,
		corsCfg:        corsCfg,
	}
}

// Routes chi http routs configuration
func (s *Handlers) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.corsCfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   s.corsCfg.AllowedHeaders,
		AllowCredentials: false,
		MaxAge:           s.corsCfg.MaxAge,
	}))

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, struct {
			Status string `json:"status"`
		}{Status: "up and running"})
	})
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/", s.proxyHandler.Send)
		api.Get("/public", s.keyHandler.GetPublicKey)

		api.With(s.authmiddleware).
			Get("/all", s.proxyHandler.GetAllMessagesByUniqueID)

		api.Get("/{id}", s.proxyHandler.Get)
		api.Post("/{id}/ack", s.proxyHandler.AckMessage)

		api.With(s.authmiddleware).
			Get("/subscribe", s.proxyHandler.SubscribeNotifications)
	})
	r.Route("/api/v2", func(api chi.Router) {
		api.Get("/{id}", s.proxyHandler.GetV2)
	})

	return r
}
