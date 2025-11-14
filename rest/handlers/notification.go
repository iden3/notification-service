package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/iden3/notification-service/log"
	"github.com/iden3/notification-service/rest/middleware"
	"github.com/iden3/notification-service/rest/utils"
	"github.com/iden3/notification-service/services"
)

// PushNotificationHandler for sending and fetching push notifications
type PushNotificationHandler struct {
	notificationService notificationService
	cachingService      cachingService
	subscriptionService subscriptionService
	pingTickerTime      time.Duration
}
type notificationService interface {
	SendNotification(ctx context.Context, msg *services.PushNotification) []services.NotificationResult
}
type cachingService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	GetAllByPrefix(ctx context.Context, prefix string) (values []interface{}, keys []string, err error)
	Delete(ctx context.Context, keys ...string) error
}

type subscriptionService interface {
	Subscribe(userDID string) (<-chan services.NotificationPayload, error)
	Unsubscribe(userDID string, uch <-chan services.NotificationPayload)
}

// NewPushNotificationHandler create new instance of proxy
func NewPushNotificationHandler(
	s notificationService,
	cs cachingService,
	sub subscriptionService,
	pingTickerTime time.Duration,
) *PushNotificationHandler {
	return &PushNotificationHandler{
		notificationService: s,
		cachingService:      cs,
		subscriptionService: sub,
		pingTickerTime:      pingTickerTime,
	}
}

// Send proxy notification to matrix sygnal gateway
func (h *PushNotificationHandler) Send(w http.ResponseWriter, r *http.Request) {
	var cReq services.PushNotification
	if err := render.DecodeJSON(r.Body, &cReq); err != nil {
		utils.ErrorJSON(w, r, http.StatusBadRequest, err, "can't bind request", 0)
		return
	}
	if err := cReq.Validate(); err != nil {
		utils.ErrorJSON(w, r, http.StatusBadRequest, err, "invalid request", 0)
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
	msg, ok := resp.(string)
	if !ok {
		utils.ErrorJSON(w, r, http.StatusNotFound, errors.New("invalid message from redis"), "error", 0)
		return
	}
	render.Status(r, http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(json.RawMessage(msg)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// delete key if everything is ok
	if err := h.cachingService.Delete(r.Context(), idParam); err != nil {
		log.Error("failed to delete key:", err, "key:", idParam)
	}
}

// Get all message by uniqueID
func (h *PushNotificationHandler) GetAllMessagesByUniqueID(w http.ResponseWriter, r *http.Request) {
	d, ok := middleware.GetDIDFromContext(r.Context())
	if !ok || d.String() == "" {
		utils.ErrorJSON(w, r, http.StatusBadRequest, errors.New("no uniqueID in context"), "can't get uniqueID from context", 0)
		return
	}
	uniqueID := d.String()

	values, keys, err := h.cachingService.GetAllByPrefix(r.Context(), uniqueID)
	if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError,
			err, "failed to get notifications", 0)
		return
	}

	if values == nil {
		utils.ErrorJSON(w, r, http.StatusNotFound,
			errors.New("notifications not found"), "expired", 0)
		return
	}

	if len(values) != len(keys) {
		utils.ErrorJSON(w, r, http.StatusInternalServerError,
			errors.New("invalid cache state"), "invalid cache state", 0)
		return
	}

	type payload struct {
		ID   string          `json:"id"`
		Body json.RawMessage `json:"body"`
	}

	respStr := make([]payload, 0, len(keys))
	for i := range keys {
		msg, ok := values[i].(string)
		if !ok {
			utils.ErrorJSON(w, r, http.StatusNotFound,
				errors.New("invalid message from redis"), "error", 0)
			return
		}
		respStr = append(respStr, payload{keys[i], json.RawMessage(msg)})
	}

	render.Status(r, http.StatusOK)
	// here we can't use render.JSON because we have to handle error in another way
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(respStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// delete keys if everything is ok
	// TODO (illia-korotia): we have to implement ACK message to remove notification
	// or manage notifications on the sender side
	// if err := h.cachingService.Delete(r.Context(), keys...); err != nil {
	// 	log.Error("failed to delete keys:", err, "keys:", keys)
	// }
}

func (h *PushNotificationHandler) SubscribeNotifications(w http.ResponseWriter, r *http.Request) {
	d, ok := middleware.GetDIDFromContext(r.Context())
	if !ok || d.String() == "" {
		utils.ErrorJSON(w, r, http.StatusBadRequest,
			errors.New("no userDID in context"), "can't get userDID from context", 0)
		return
	}
	userDID := d.String()

	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.ErrorJSON(w, r, http.StatusInternalServerError,
			errors.New("streaming unsupported"), "streaming unsupported", 0)
		return
	}

	// since HTTP2 doesn't have limitation of open connections per client,
	// we limit number of open subscriptions per userDID to prevent memory leak
	ch, err := h.subscriptionService.Subscribe(userDID)
	if err != nil && errors.Is(err, services.ErrMaxSubscriptionsReached) {
		utils.ErrorJSON(w, r, http.StatusTooManyRequests,
			err, "maximum number of open subscriptions reached", 0)
		return
	} else if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError,
			err, "failed to subscribe to notifications", 0)
		return
	}

	defer h.subscriptionService.Unsubscribe(userDID, ch)

	pingTicker := time.NewTicker(h.pingTickerTime)
	defer pingTicker.Stop()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				// unsubscribe channel closed
				_, _ = fmt.Fprint(w, utils.CloseMessage)
				flusher.Flush()
				return
			}

			event := utils.BuildEventMessage(data)
			_, _ = fmt.Fprint(w, event)
			flusher.Flush()
		case <-pingTicker.C:
			_, _ = fmt.Fprint(w, utils.PingMessage)
			flusher.Flush()
		case <-r.Context().Done():
			log.WithContext(r.Context()).Info(
				"connection closed",
				slog.String(
					"reason",
					"context done",
				),
			)
			return
		}
	}
}
