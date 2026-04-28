package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// Analytics serves the dashboard's summary endpoints.
type Analytics struct {
	Repo *store.Analytics
	Log  *slog.Logger
}

// Overview handles GET /api/v1/analytics/overview — KPIs + carrier ranking.
func (h *Analytics) Overview(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	overview, err := h.Repo.FetchOverview(ctx)
	if err != nil {
		h.Log.Error("overview failed", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load overview")
		return
	}

	carriers, err := h.Repo.FetchCarriers(ctx, 5)
	if err != nil {
		h.Log.Warn("carriers failed", "err", err)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"overview": overview,
		"carriers": carriers,
	})
}
