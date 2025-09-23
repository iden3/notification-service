package handlers

import (
	"net/http"

	"github.com/iden3/notification-service/log"
	"github.com/iden3/notification-service/rest/utils"
	"github.com/iden3/notification-service/services"
)

// KeyHandler is a handler for ppg key info
type KeyHandler struct {
	keyService *services.Crypto
}

// NewKeyHandler creates new handler for public key queries
func NewKeyHandler(s *services.Crypto) *KeyHandler {
	return &KeyHandler{keyService: s}
}

// GetPublicKey return public key of push gateway
func (h *KeyHandler) GetPublicKey(w http.ResponseWriter, r *http.Request) {
	pkPem, err := h.keyService.MarshalPubKeyToPem()
	if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed encode public key", 0)
		return
	}
	_, err = w.Write(pkPem)
	if err != nil {
		log.Warn("failed write response:", err)
	}
}
