package handler

import (
	"net/http"

	"classified-vault/internal/domain"
)

type AuditService interface {
	List(limit int) ([]*domain.AuditLog, error)
}

type AuditHandler struct {
	service AuditService
}

func NewAuditHandler(service AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	logs, err := h.service.List(50)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list audit logs"})
		return
	}

	writeJSON(w, http.StatusOK, logs)
}
