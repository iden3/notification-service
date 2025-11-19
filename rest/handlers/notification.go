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
	expirationDuration  time.Duration
}
type notificationService interface {
	SendNotification(ctx context.Context, msg *services.PushNotification) []services.NotificationResult
}
type cachingService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	GetAllByPrefix(ctx context.Context, prefix string) (values []interface{}, keys []string, err error)
	Delete(ctx context.Context, keys ...string) error
	Set(ctx context.Context, key string, value interface{}, duration time.Duration) error
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
	expirationDuration time.Duration,
) *PushNotificationHandler {
	return &PushNotificationHandler{
		notificationService: s,
		cachingService:      cs,
		subscriptionService: sub,
		pingTickerTime:      pingTickerTime,
		expirationDuration:  expirationDuration,
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
// returns only body to keep backward compatibility
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
	respStr, ok := resp.(string)
	if !ok {
		utils.ErrorJSON(w, r, http.StatusNotFound, errors.New("invalid message from redis"), "error", 0)
		return
	}

	var msg services.NotificationContent
	if err := json.Unmarshal([]byte(respStr), &msg); err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed to unmarshal notification", 0)
		return
	}

	// old message format doesn't contain metadata
	var payload []byte
	if services.IsEmptyMetadata(msg.Metadata) {
		// it true: message has raw format without metadata
		payload = []byte(respStr)
	} else {
		// if false: message has new format with metadata
		payload = msg.Body
	}

	render.Status(r, http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// return body only to keep backward compatibility with old clients
	if err := json.NewEncoder(w).Encode(json.RawMessage(payload)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetV2 returns notification by identifier for v2 API
// returns body and metadata
func (h *PushNotificationHandler) GetV2(w http.ResponseWriter, r *http.Request) {
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
	respStr, ok := resp.(string)
	if !ok {
		utils.ErrorJSON(w, r, http.StatusNotFound, errors.New("invalid message from redis"), "error", 0)
		return
	}

	render.Status(r, http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(json.RawMessage(respStr)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		ID       string                        `json:"id"`
		Body     json.RawMessage               `json:"body"`
		Metadata services.NotificationMetadata `json:"metadata"`
	}

	respStr := make([]payload, 0, len(keys))
	for i := range keys {
		msg, ok := values[i].(string)
		if !ok {
			utils.ErrorJSON(w, r, http.StatusNotFound,
				errors.New("invalid message from redis"), "error", 0)
			return
		}

		var nContent services.NotificationContent
		if err := json.Unmarshal([]byte(msg), &nContent); err != nil {
			utils.ErrorJSON(w, r, http.StatusInternalServerError,
				err, "failed to unmarshal notification", 0)
			return
		}

		// TODO (illia-korotia): this is temporary fix for old messages without metadata
		// when all old messages will be expired, we can remove this block
		body := nContent.Body
		if services.IsEmptyMetadata(nContent.Metadata) {
			// old message format without metadata
			body = []byte(msg)
		}

		respStr = append(respStr, payload{
			ID:       keys[i],
			Body:     body,
			Metadata: nContent.Metadata,
		})
	}

	render.Status(r, http.StatusOK)
	// here we can't use render.JSON because we have to handle error in another way
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(respStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// AckMessage marks a message as read
func (h *PushNotificationHandler) AckMessage(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		utils.ErrorJSON(w, r, http.StatusBadRequest, errors.New("no id param"), "can't get notification id param", 0)
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

	respStr, ok := resp.(string)
	if !ok {
		utils.ErrorJSON(w, r, http.StatusNotFound, errors.New("invalid message from redis"), "error", 0)
		return
	}

	var msg services.NotificationContent
	if err := json.Unmarshal([]byte(respStr), &msg); err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed to unmarshal notification", 0)
		return
	}

	msg.Metadata.ReadAt = time.Now().UTC()
	msg.Metadata.IsRead = true

	updatedBytes, err := json.Marshal(msg)
	if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed to marshal updated notification", 0)
		return
	}

	err = h.cachingService.Set(r.Context(), idParam, updatedBytes, h.expirationDuration)
	if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed to update notification", 0)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
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
