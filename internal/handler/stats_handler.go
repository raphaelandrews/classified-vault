package handler

import (
	"net/http"

	"classified-vault/internal/repository"
)

type StatsService interface {
	GetStats() (*repository.StatsResponse, error)
}

type StatsHandler struct {
	service StatsService
}

func NewStatsHandler(service StatsService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	session := requireSession(w, r)
	if session == nil {
		return
	}

	stats, err := h.service.GetStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get stats"})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
