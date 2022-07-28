package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/iden3/notification-service/services"
	"net/http"

	"github.com/iden3/notification-service/log"
	"github.com/iden3/notification-service/rest/utils"
)

// PushNotificationHandler for sending and fetching push notifications
type PushNotificationHandler struct {
	notificationService notificationService
	cachingService      cachingService
}
type notificationService interface {
	SendNotification(ctx context.Context, msg *services.Message) []services.NotificationResult
}
type cachingService interface {
	Get(ctx context.Context, key string) (interface{}, error)
}

// NewPushNotificationHandler create new instance of proxy
func NewPushNotificationHandler(s notificationService, cs cachingService) *PushNotificationHandler {
	return &PushNotificationHandler{notificationService: s, cachingService: cs}
}

// Send proxy notification to matrix sygnal gateway
func (h *PushNotificationHandler) Send(w http.ResponseWriter, r *http.Request) {
	var cReq services.Message
	if err := render.DecodeJSON(r.Body, &cReq); err != nil {
		utils.ErrorJSON(w, r, http.StatusBadRequest, err, "can't bind request", 0)
		return
	}

	resp := h.notificationService.SendNotification(r.Context(), &cReq)

	respBytes, err := json.Marshal(resp)
	if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed proxy notification", 0)
		return
	}
	w.Header().Add("Content-type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		log.Warn("failed write response:", err)
	}
}

// Get returns notification by identifier
func (h *PushNotificationHandler) Get(w http.ResponseWriter, r *http.Request) {

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		utils.ErrorJSON(w, r, http.StatusBadRequest, errors.New("no id param"), "can't  get notification id param", 0)
		return
	}

	resp, err := h.cachingService.Get(r.Context(), idParam)
	if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed to get notification", 0)
		return
	}
	if resp == nil {
		utils.ErrorJSON(w, r, http.StatusNotFound, errors.New("notification not found"), "expired", 0)
		return
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed to pack notification message", 0)
		return
	}
	w.Header().Add("Content-type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		log.Warn("failed write response:", err)
	}
}
