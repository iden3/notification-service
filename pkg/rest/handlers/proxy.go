package handlers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/iden3/signed-notification/pkg/log"
	"github.com/iden3/signed-notification/pkg/rest/utils"
	"github.com/iden3/signed-notification/pkg/service"
)

type ProxyHandler struct {
	proxyService *service.ProxyService
}

func NewProxyHandler(s *service.ProxyService) *ProxyHandler {
	return &ProxyHandler{proxyService: s}
}

func (h *ProxyHandler) ProxyNotification(w http.ResponseWriter, r *http.Request) {
	var cReq service.Message
	if err := render.DecodeJSON(r.Body, &cReq); err != nil {
		utils.ErrorJSON(w, r, http.StatusBadRequest, err, "can't bind request", 0)
		return
	}

	resp, err := h.proxyService.Proxy(r.Context(), &cReq)
	if err != nil {
		utils.ErrorJSON(w, r, http.StatusInternalServerError, err, "failed proxy notification", 0)
		return
	}

	w.Header().Add("Content-type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		log.Warn("failed write response:", err)
	}
}

