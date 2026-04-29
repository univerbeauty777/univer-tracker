package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// Gargalos serves the rastreiaki gargalos dashboards.
type Gargalos struct {
	Repo *store.Transitions
	Log  *slog.Logger
}

// Funnel handles GET /api/v1/analytics/funnel.
func (h *Gargalos) Funnel(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	out, err := h.Repo.Funnel(ctx, days)
	if err != nil {
		h.Log.Error("funnel failed", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load funnel")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"stages": out})
}

// Transitions handles GET /api/v1/analytics/transitions — full table for
// the gargalos page.
func (h *Gargalos) Transitions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	out, err := h.Repo.All(ctx, days)
	if err != nil {
		h.Log.Error("transitions failed", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load transitions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"transitions": out})
}
