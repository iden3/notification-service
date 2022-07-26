package handlers

import (
	"net/http"

	"github.com/iden3/signed-notification/pkg/log"
	"github.com/iden3/signed-notification/pkg/rest/utils"
	"github.com/iden3/signed-notification/pkg/service"
)

type KeyHandler struct {
	keyService *service.Cryptographer
}

func NewKeyHandler(s *service.Cryptographer) *KeyHandler {
	return &KeyHandler{keyService: s}
}

func (h *KeyHandler) GetPublicKey(w http.ResponseWriter, r *http.Request) {
	pkPem, err := h.keyService.MarshalToPemPublicKey()
	if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed read public key", 0)
		return
	}
	_, err = w.Write(pkPem)
	if err != nil {
		log.Warn("failed write response:", err)
	}
}
