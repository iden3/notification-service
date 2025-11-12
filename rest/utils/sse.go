package utils

import (
	"encoding/json"
	"log/slog"

	"github.com/iden3/notification-service/log"
	"github.com/iden3/notification-service/services"
)

// BuildEventMessage builds SSE event message for new notifications
func BuildEventMessage(payload services.NotificationPayload) string {
	event := "event: new_notifications\n"
	message := "data: "
	d, err := json.Marshal(payload)
	if err != nil {
		log.Error("Error marshaling new notifications", slog.String("error", err.Error()))
		return message + "{}\n\n"
	}
	return event + message + string(d) + "\n\n"
}

// PingMessage builds SSE ping message
const PingMessage = "event: ping\ndata: {}\n\n"

// CloseMessage builds SSE close message
const CloseMessage = "event: close\ndata: {\"reason\":\"unsubscribed\"}\n\n"
